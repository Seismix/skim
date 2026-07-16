// skim — a tiny, single-binary social-preview viewer.
//
// It runs a local HTTP server that (1) serves the embedded UI and
// (2) fetches a URL *on this machine* and extracts its OpenGraph /
// Twitter / standard meta tags. Because the fetch happens locally,
// it can inspect http://localhost:* dev servers — the whole point.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// The Svelte UI is built by `vite build` (see Makefile) into web/dist and
// embedded here. Build the frontend before `go build` — `make` does both.
//
//go:embed all:web/dist
var uiFS embed.FS

// defaultUA mimics the canonical OpenGraph crawler. Sites commonly serve their
// share metadata only to recognized crawlers (Reddit, for example, returns an
// anti-bot landing page otherwise), so this makes skim see what Facebook /
// WhatsApp / Discord / Slack actually see. Override with --user-agent.
const defaultUA = "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)"

var userAgent = defaultUA

// version is stamped at build time via -ldflags "-X main.version=...". When unset
// (e.g. `go install ...@v1.2.3`), it falls back to the module version from the
// embedded build info — see buildVersion.
var version = "dev"

// buildVersion resolves the version to print for -v. A linker-stamped version
// wins; otherwise we read the module version baked in by `go install pkg@ver`.
func buildVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	port := flag.Int("port", 0, "port to listen on (0 = pick a free one)")
	noOpen := flag.Bool("no-open", false, "do not open the browser on launch")
	quiet := flag.Bool("quiet", false, "suppress startup/shutdown banners (used by the dev server)")
	uaFlag := flag.String("user-agent", defaultUA, "User-Agent for outbound fetches (default mimics the OpenGraph crawler)")
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "print the version and exit")
	flag.BoolVar(&showVersion, "version", false, "print the version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("skim %s\n", buildVersion())
		return
	}

	userAgent = *uaFlag

	// Optional positional arg: a URL to pre-fill and auto-preview.
	presetURL := ""
	if args := flag.Args(); len(args) > 0 {
		presetURL = args[0]
	}

	sub, err := fs.Sub(uiFS, "web/dist")
	if err != nil {
		fatal("embed: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/fetch-og", handleFetchOG)
	mux.HandleFunc("/api/img", handleImg)
	mux.Handle("/", http.FileServer(http.FS(sub)))

	// skim is a personal, local tool — always bind loopback.
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		fatal("listen: %v", err)
	}

	addr := ln.Addr().(*net.TCPAddr)
	openURL := fmt.Sprintf("http://127.0.0.1:%d/", addr.Port)
	if presetURL != "" {
		openURL += "?url=" + url.QueryEscape(presetURL)
	}

	if !*quiet {
		fmt.Printf("skim is running at %s\n", openURL)
		fmt.Println("press Ctrl+C to stop")
	}
	if !*noOpen {
		if err := openBrowser(openURL); err != nil {
			fmt.Printf("(could not open browser automatically: %v)\n", err)
		}
	}

	// Stop cleanly on Ctrl+C / SIGTERM so the listener closes and the port frees.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			fatal("serve: %v", err)
		}
	}()

	<-ctx.Done()
	if !*quiet {
		fmt.Println("\nshutting down — releasing port")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

// ogData mirrors the JSON shape the UI expects
type ogData struct {
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Image          string            `json:"image"`
	ImageWidth     string            `json:"imageWidth"`
	ImageHeight    string            `json:"imageHeight"`
	ImageAlt       string            `json:"imageAlt"`
	URL            string            `json:"url"`
	SiteName       string            `json:"siteName"`
	Type           string            `json:"type"`
	TwitterCard    string            `json:"twitterCard"`
	TwitterSite    string            `json:"twitterSite"`
	TwitterCreator string            `json:"twitterCreator"`
	Locale         string            `json:"locale"`
	ThemeColor     string            `json:"themeColor"`
	Favicon        string            `json:"favicon"`
	AllMeta        map[string]string `json:"allMeta"`
}

// fetchResult is one entry of a batch response: the input URL paired with its
// extracted data or a per-URL error, so one bad link never fails the whole set.
type fetchResult struct {
	URL   string  `json:"url"`
	Data  *ogData `json:"data,omitempty"`
	Error string  `json:"error,omitempty"`
}

func handleFetchOG(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Accept either a single "url" (legacy shape, returns one ogData) or a "urls"
	// array (batch shape, returns {"results":[...]}). The UI sends "urls"; the
	// single form stays for shared links and existing API callers.
	var body struct {
		URL  string   `json:"url"`
		URLs []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if len(body.URLs) > 0 {
		// Fetch every URL concurrently; preserve input order via an indexed slice.
		results := make([]fetchResult, len(body.URLs))
		var wg sync.WaitGroup
		for i, u := range body.URLs {
			results[i].URL = u
			if strings.TrimSpace(u) == "" {
				results[i].Error = "URL is required"
				continue
			}
			wg.Add(1)
			go func(i int, u string) {
				defer wg.Done()
				data, err := skim(r.Context(), u)
				if err != nil {
					results[i].Error = err.Error()
					return
				}
				results[i].Data = data
			}(i, u)
		}
		wg.Wait()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
		return
	}

	if body.URL == "" {
		writeErr(w, http.StatusBadRequest, "URL is required")
		return
	}
	data, err := skim(r.Context(), body.URL)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

// skim fetches rawURL on this machine and extracts its OpenGraph / Twitter /
// standard meta tags into an ogData. The returned error is safe to surface to
// the caller (it carries no internal state).
func skim(ctx context.Context, rawURL string) (*ogData, error) {
	// Bare hosts get a scheme: http for localhost, https for everything else.
	target := normalizeURL(rawURL)
	parsed, err := url.Parse(target)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("only HTTP/HTTPS URLs are allowed")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	meta := map[string]string{}
	doc.Find("meta").Each(func(_ int, s *goquery.Selection) {
		key, ok := s.Attr("property")
		if !ok || key == "" {
			key, _ = s.Attr("name")
		}
		content, _ := s.Attr("content")
		if key != "" && content != "" {
			meta[strings.ToLower(key)] = content
		}
	})

	// rel may be "icon", "shortcut icon", or "icon shortcut" — match any listing "icon".
	favicon, _ := doc.Find(`link[rel~="icon"]`).Attr("href")

	data := ogData{
		Title:          firstNonEmpty(meta["og:title"], meta["twitter:title"], strings.TrimSpace(doc.Find("title").First().Text())),
		Description:    firstNonEmpty(meta["og:description"], meta["twitter:description"], meta["description"]),
		Image:          firstNonEmpty(meta["og:image"], meta["twitter:image"]),
		ImageWidth:     meta["og:image:width"],
		ImageHeight:    meta["og:image:height"],
		ImageAlt:       firstNonEmpty(meta["og:image:alt"], meta["twitter:image:alt"]),
		URL:            firstNonEmpty(meta["og:url"], target),
		SiteName:       meta["og:site_name"],
		Type:           meta["og:type"],
		TwitterCard:    meta["twitter:card"],
		TwitterSite:    meta["twitter:site"],
		TwitterCreator: meta["twitter:creator"],
		Locale:         meta["og:locale"],
		ThemeColor:     meta["theme-color"],
		Favicon:        favicon,
		AllMeta:        meta,
	}

	// Resolve relative image / favicon URLs against the fetched page.
	data.Image = resolveURL(parsed, data.Image)
	data.Favicon = resolveURL(parsed, data.Favicon)

	return &data, nil
}

// handleImg proxies an external image through this server so it loads as a
// same-origin resource — bypassing browser tracking protection, hotlink
// referer checks, and mixed-content rules that otherwise block og:images.
func handleImg(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("url")
	parsed, err := url.Parse(raw)
	if raw == "" || err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "image/*,*/*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "private, max-age=300")
	_, _ = io.Copy(w, io.LimitReader(resp.Body, 15<<20))
}

var schemeRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.\-]*://`)

// normalizeURL adds a scheme to bare input. localhost-style hosts default to
// http:// (dev servers); everything else defaults to https://.
func normalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || schemeRe.MatchString(raw) {
		return raw
	}
	host := raw
	if i := strings.IndexAny(host, "/?#"); i >= 0 {
		host = host[:i]
	}
	// Strip the port, but keep bracketed IPv6 hosts ("[::1]", "[::1]:8080") whole —
	// LastIndex(":") would otherwise chop inside the address.
	if strings.HasPrefix(host, "[") {
		if j := strings.IndexByte(host, ']'); j >= 0 {
			host = host[:j+1]
		}
	} else if i := strings.LastIndex(host, ":"); i >= 0 {
		host = host[:i] // strip port
	}
	if isLocalhost(strings.ToLower(host)) {
		return "http://" + raw
	}
	return "https://" + raw
}

func isLocalhost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "0.0.0.0", "::1", "[::1]":
		return true
	}
	return strings.HasSuffix(host, ".localhost")
}

func resolveURL(base *url.URL, ref string) string {
	if ref == "" || strings.HasPrefix(ref, "http") {
		return ref
	}
	r, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(r).String()
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func openBrowser(target string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd, args = "cmd", []string{"/c", "start", "", target}
	case "darwin":
		cmd, args = "open", []string{target}
	default:
		cmd, args = "xdg-open", []string{target}
	}
	return exec.Command(cmd, args...).Start()
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "skim: "+format+"\n", a...)
	os.Exit(1)
}
