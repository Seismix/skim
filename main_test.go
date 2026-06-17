package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
