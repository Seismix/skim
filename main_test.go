package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"bare host gets https", "reddit.com", "https://reddit.com"},
		{"localhost gets http", "localhost", "http://localhost"},
		{"localhost with port gets http", "localhost:3000", "http://localhost:3000"},
		{"127.0.0.1 gets http", "127.0.0.1:8080", "http://127.0.0.1:8080"},
		{"loopback ipv6 with port gets http", "[::1]:8080", "http://[::1]:8080"},
		{"bare loopback ipv6 gets http", "[::1]", "http://[::1]"},
		{".localhost subdomain gets http", "app.localhost:5173", "http://app.localhost:5173"},
		{"existing https scheme untouched", "https://example.com", "https://example.com"},
		{"existing http scheme untouched", "http://localhost:3000", "http://localhost:3000"},
		{"path and query preserved", "example.com/path?q=1", "https://example.com/path?q=1"},
		{"surrounding whitespace trimmed", "  reddit.com  ", "https://reddit.com"},
		{"empty stays empty", "", ""},
		{"uppercase localhost host detected", "LOCALHOST:3000", "http://LOCALHOST:3000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeURL(tt.in); got != tt.want {
				t.Errorf("normalizeURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsLocalhost(t *testing.T) {
	local := []string{"localhost", "127.0.0.1", "0.0.0.0", "::1", "[::1]", "app.localhost"}
	remote := []string{"example.com", "reddit.com", "127.0.0.1.evil.com", "notlocalhost", ""}

	for _, h := range local {
		if !isLocalhost(h) {
			t.Errorf("isLocalhost(%q) = false, want true", h)
		}
	}
	for _, h := range remote {
		if isLocalhost(h) {
			t.Errorf("isLocalhost(%q) = true, want false", h)
		}
	}
}

func TestHandleFetchOG(t *testing.T) {
	const html = `<!doctype html><html><head>
<title>Fallback Title</title>
<meta property="og:title" content="OG Title">
<meta name="twitter:title" content="Twitter Title">
<meta property="og:description" content="OG Desc">
<meta property="og:image" content="/img/card.png">
<meta property="og:image:width" content="1200">
<meta name="twitter:card" content="summary_large_image">
<link rel="icon" href="/favicon.ico">
</head><body>hi</body></html>`

	data, base := fetchOG(t, html)

	if data.Title != "OG Title" {
		t.Errorf("Title = %q, want %q (og:title should win over twitter:title and <title>)", data.Title, "OG Title")
	}
	if data.Description != "OG Desc" {
		t.Errorf("Description = %q, want %q", data.Description, "OG Desc")
	}
	if want := base + "/img/card.png"; data.Image != want {
		t.Errorf("Image = %q, want %q (relative og:image should resolve against the page)", data.Image, want)
	}
	if data.ImageWidth != "1200" {
		t.Errorf("ImageWidth = %q, want %q", data.ImageWidth, "1200")
	}
	if data.TwitterCard != "summary_large_image" {
		t.Errorf("TwitterCard = %q, want %q", data.TwitterCard, "summary_large_image")
	}
	if want := base + "/favicon.ico"; data.Favicon != want {
		t.Errorf("Favicon = %q, want %q (relative favicon should resolve against the page)", data.Favicon, want)
	}
	if data.ImageKind != "banner" {
		t.Errorf("ImageKind = %q, want %q (a 1200-wide og:image classifies as a banner)", data.ImageKind, "banner")
	}
}

func TestImageKind(t *testing.T) {
	tests := []struct {
		name string
		d    ogData
		want string
	}{
		{"no image", ogData{}, "none"},
		{"favicon as og:image", ogData{Image: "https://x.com/favicon.ico"}, "icon"},
		{"real banner", ogData{Image: "https://x.com/card.png", ImageWidth: "1200"}, "banner"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := imageKind(&tt.d); got != tt.want {
				t.Errorf("imageKind = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandleFetchOGTitleFallback(t *testing.T) {
	// No og:title — twitter:title should win over the <title> element.
	const twHTML = `<html><head><title>Doc Title</title>
<meta name="twitter:title" content="Twitter Title"></head><body></body></html>`
	if data, _ := fetchOG(t, twHTML); data.Title != "Twitter Title" {
		t.Errorf("Title = %q, want %q (twitter:title should win over <title>)", data.Title, "Twitter Title")
	}

	// Neither og:title nor twitter:title — fall back to the <title> element.
	const titleHTML = `<html><head><title>Doc Title</title></head><body></body></html>`
	if data, _ := fetchOG(t, titleHTML); data.Title != "Doc Title" {
		t.Errorf("Title = %q, want %q (should fall back to <title>)", data.Title, "Doc Title")
	}
}

func TestHandleFetchOGBatch(t *testing.T) {
	const html = `<html><head><meta property="og:title" content="Batch Title"></head><body></body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, html)
	}))
	defer srv.Close()

	// One reachable URL, one empty entry: the batch should report a result per
	// input in order, the empty one carrying an error rather than failing the set.
	body := strings.NewReader(`{"urls":[` + strconv.Quote(srv.URL) + `,""]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/fetch-og", body)
	rec := httptest.NewRecorder()
	handleFetchOG(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var out struct {
		Results []fetchResult `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(out.Results) != 2 {
		t.Fatalf("got %d results, want 2", len(out.Results))
	}
	if out.Results[0].Data == nil || out.Results[0].Data.Title != "Batch Title" {
		t.Errorf("results[0] = %+v, want data with Title %q", out.Results[0], "Batch Title")
	}
	if out.Results[1].Error == "" {
		t.Errorf("results[1] error = %q, want a non-empty error for the empty URL", out.Results[1].Error)
	}
}

func TestSelectPlatforms(t *testing.T) {
	// Absent means every platform; an explicit empty array means none. That split is
	// the whole reason the field is a pointer, so it's worth pinning down.
	if got, err := selectPlatforms(nil); err != nil || len(got) != len(platformIDs) {
		t.Errorf("selectPlatforms(nil) = %v, %v; want all %d platforms", got, err, len(platformIDs))
	}
	if got, err := selectPlatforms(&[]string{}); err != nil || len(got) != 0 {
		t.Errorf("selectPlatforms([]) = %v, %v; want an empty selection", got, err)
	}

	got, err := selectPlatforms(&[]string{" IMESSAGE ", "slack", "imessage"})
	if err != nil {
		t.Fatalf("selectPlatforms: unexpected error %v", err)
	}
	if want := []string{"imessage", "slack"}; !slices.Equal(got, want) {
		t.Errorf("selectPlatforms = %v, want %v (trimmed, lowercased, deduped)", got, want)
	}

	if _, err := selectPlatforms(&[]string{"myspace"}); err == nil {
		t.Error("selectPlatforms(myspace) = nil error, want a rejection naming the valid platforms")
	}
}

func TestIsIconImage(t *testing.T) {
	tests := []struct {
		name  string
		image string
		width string
		want  bool
	}{
		{"no image is not an icon", "", "", false},
		{"favicon in the name", "https://x.com/favicon.ico", "", true},
		{"apple-touch icon", "https://x.com/apple-touch-icon.png", "", true},
		{"icon path segment", "https://x.com/icon/logo.jpg", "", true},
		{"narrow image is an icon", "https://x.com/card.png", "96", true},
		{"width with a unit still parses", "https://x.com/card.png", "96px", true},
		{"wide image is a banner", "https://x.com/card.png", "1200", false},
		{"exactly 200 is a banner", "https://x.com/card.png", "200", false},
		{"junk width is not an icon", "https://x.com/card.png", "wide", false},
		{"no width is not an icon", "https://x.com/card.png", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ogData{Image: tt.image, ImageWidth: tt.width}
			if got := isIconImage(d); got != tt.want {
				t.Errorf("isIconImage(%q, w=%q) = %v, want %v", tt.image, tt.width, got, tt.want)
			}
		})
	}
}

func TestRenderPlatformsShapes(t *testing.T) {
	// The three shapes WhatsApp and iMessage pick between, driven by the image.
	tests := []struct {
		name     string
		d        ogData
		whatsapp string
		imessage string
		twitter  string
	}{
		{
			name:     "banner image",
			d:        ogData{Image: "https://x.com/card.png", ImageWidth: "1200", TwitterCard: "summary_large_image"},
			whatsapp: "bubble · 1.91 : 1",
			imessage: "rich link · banner",
			twitter:  "summary_large_image · 2 : 1",
		},
		{
			name:     "icon-sized image demotes to a thumb",
			d:        ogData{Image: "https://x.com/favicon.ico"},
			whatsapp: "bubble · thumb",
			imessage: "rich link · thumb",
			twitter:  "summary · 1 : 1",
		},
		{
			name:     "no image at all",
			d:        ogData{},
			whatsapp: "bubble",
			imessage: "rich link",
			twitter:  "summary · 1 : 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderPlatforms(&tt.d, platformIDs)
			if got["whatsapp"].Shape != tt.whatsapp {
				t.Errorf("whatsapp shape = %q, want %q", got["whatsapp"].Shape, tt.whatsapp)
			}
			if got["imessage"].Shape != tt.imessage {
				t.Errorf("imessage shape = %q, want %q", got["imessage"].Shape, tt.imessage)
			}
			if got["twitter"].Shape != tt.twitter {
				t.Errorf("twitter shape = %q, want %q", got["twitter"].Shape, tt.twitter)
			}
		})
	}
}

func TestRenderPlatformsFields(t *testing.T) {
	d := ogData{
		Title:       "Title",
		Description: "Desc",
		Image:       "https://x.com/card.png",
		ImageWidth:  "1200",
		URL:         "https://x.com/post",
		SiteName:    "Site",
	}
	got := renderPlatforms(&d, platformIDs)

	// null vs. empty is the contract: platforms that never show a field report null,
	// so a caller can tell "ignored" from "you left it blank".
	if got["imessage"].Description != nil {
		t.Errorf("imessage description = %v, want null (iMessage ignores og:description)", *got["imessage"].Description)
	}
	if got["linkedin"].Description != nil {
		t.Errorf("linkedin description = %v, want null (the card shows title + domain only)", *got["linkedin"].Description)
	}
	if got["facebook"].Description == nil || *got["facebook"].Description != "Desc" {
		t.Errorf("facebook description = %v, want %q", got["facebook"].Description, "Desc")
	}
	if got["facebook"].Domain != "x.com" {
		t.Errorf("facebook domain = %q, want %q (hostname of og:url)", got["facebook"].Domain, "x.com")
	}
	if got["slack"].SiteName == nil || *got["slack"].SiteName != "Site" {
		t.Errorf("slack siteName = %v, want %q", got["slack"].SiteName, "Site")
	}

	// Slack's header falls back to the domain when og:site_name is missing.
	bare := ogData{URL: "https://x.com/post"}
	if s := renderPlatforms(&bare, []string{"slack"})["slack"]; s.SiteName == nil || *s.SiteName != "x.com" {
		t.Errorf("slack siteName without og:site_name = %v, want the domain %q", s.SiteName, "x.com")
	}

	// An icon-sized og:image is a provider icon on Discord/Slack, not an embed, but
	// WhatsApp still renders it as a thumb — so only the former report null.
	icon := ogData{Image: "https://x.com/favicon.ico", URL: "https://x.com/post"}
	iconGot := renderPlatforms(&icon, platformIDs)
	if iconGot["discord"].Image != nil {
		t.Errorf("discord image = %v, want null for an icon-sized og:image", *iconGot["discord"].Image)
	}
	if iconGot["slack"].Image != nil {
		t.Errorf("slack image = %v, want null for an icon-sized og:image", *iconGot["slack"].Image)
	}
	if iconGot["whatsapp"].Image == nil {
		t.Error("whatsapp image = null, want the icon (WhatsApp renders it as a thumb)")
	}
}

func TestHandleFetchOGPlatformsParam(t *testing.T) {
	const html = `<html><head><meta property="og:title" content="T"></head><body></body></html>`

	// Absent: every platform reported, so the default costs the caller nothing.
	if data, _ := fetchOG(t, html); len(data.Platforms) != len(platformIDs) {
		t.Errorf("platforms block has %d entries, want all %d when the field is omitted", len(data.Platforms), len(platformIDs))
	}

	// An unknown id is a 400, not a silent empty block.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, html)
	}))
	defer srv.Close()

	body := strings.NewReader(`{"url":` + strconv.Quote(srv.URL) + `,"platforms":["myspace"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/fetch-og", body)
	rec := httptest.NewRecorder()
	handleFetchOG(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for an unknown platform", rec.Code)
	}

	// A selection narrows the block to exactly what was asked for.
	body = strings.NewReader(`{"url":` + strconv.Quote(srv.URL) + `,"platforms":["imessage"]}`)
	req = httptest.NewRequest(http.MethodPost, "/api/fetch-og", body)
	rec = httptest.NewRecorder()
	handleFetchOG(rec, req)

	var data ogData
	if err := json.NewDecoder(rec.Body).Decode(&data); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(data.Platforms) != 1 {
		t.Fatalf("platforms block has %d entries, want 1", len(data.Platforms))
	}
	if _, ok := data.Platforms["imessage"]; !ok {
		t.Errorf("platforms block = %v, want only imessage", data.Platforms)
	}
}

// fetchOG serves html from a throwaway server, runs it through handleFetchOG,
// and returns the parsed result plus the server's base URL (for resolving the
// relative-URL assertions).
func fetchOG(t *testing.T, html string) (ogData, string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, html)
	}))
	defer srv.Close()

	body := strings.NewReader(`{"url":` + strconv.Quote(srv.URL) + `}`)
	req := httptest.NewRequest(http.MethodPost, "/api/fetch-og", body)
	rec := httptest.NewRecorder()
	handleFetchOG(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	var data ogData
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return data, srv.URL
}
