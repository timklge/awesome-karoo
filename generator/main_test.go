package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.HandlerFunc) (*metadataClient, string) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &metadataClient{
		http: server.Client(), apiBase: server.URL, token: "test-token",
		warn: func(string, ...any) {},
	}, server.URL
}

func TestGithubRepo(t *testing.T) {
	for address, want := range map[string]string{
		"https://github.com/timklge/karoo-headwind":          "timklge/karoo-headwind",
		"https://github.com/owner/repo.git/":                 "owner/repo",
		"https://github.com/owner/repo/tree/main?tab=readme": "owner/repo",
		"https://GITHUB.COM/owner/repo":                      "owner/repo",
		"https://portcity.tech/karoo-gearbeeper/":            "",
		"https://github.com/owner":                           "",
		"https://github.com.evil.example/owner/repo":         "",
		"https://user:password@github.com/owner/repo":        "",
		"http://github.com/owner/repo":                       "",
		"https://github.com/owner/repo%2Fmalicious":          "",
		"https://github.com/../repo":                         "",
	} {
		t.Run(address, func(t *testing.T) {
			if got := githubRepo(address); got != want {
				t.Errorf("githubRepo(%q) = %q, want %q", address, got, want)
			}
		})
	}
}

func TestEnrich(t *testing.T) {
	c, base := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		if r.URL.Path == "/manifest.json" {
			if r.Header.Get("Authorization") != "" {
				t.Error("GitHub token leaked to manifest server")
			}
			fmt.Fprint(w, `{
				"latestVersion":"1.8.0","description":"Manifest description",
				"iconUrl":"https://images.example.org/icon.png",
				"screenshotUrls":["https://images.example.org/preview1.png","https://images.example.org/preview2.png","https://images.example.org/preview1.png","javascript:alert(1)"]
			}`)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" || r.Header.Get("X-GitHub-Api-Version") != "2022-11-28" {
			t.Error("missing GitHub API headers")
		}
		switch r.URL.Path {
		case "/repos/owner/repo":
			fmt.Fprint(w, `{"stargazers_count":42,"license":{"spdx_id":"MIT"}}`)
		case "/repos/owner/repo/releases":
			fmt.Fprint(w, `[
				{"published_at":"2026-01-02T03:04:05Z","assets":[{"name":"app.apk","download_count":20},{"name":"manifest.json","download_count":1000}]},
				{"published_at":"2026-01-03T03:04:05Z","prerelease":true,"assets":[{"name":"preview.APK","download_count":3}]},
				{"published_at":"2026-01-04T03:04:05Z","draft":true,"assets":[{"name":"draft.apk","download_count":999}]}
			]`)
		default:
			http.NotFound(w, r)
		}
	})
	manifestURL := base + "/manifest.json"
	ext := extension{
		Name: "Test", Homepage: "https://github.com/owner/repo", Manifest: &manifestURL,
		AddedAt:       "2024-12-21",
		Compatibility: "both",
		Features:      []string{"Curated feature"}, Tags: []string{"training"}, Featured: true,
	}
	e := c.enrich(context.Background(), ext)
	if e.Version == nil || *e.Version != "1.8.0" || e.Description != "Manifest description" {
		t.Fatalf("incorrect manifest metadata: %+v", e)
	}
	if e.Stars == nil || *e.Stars != 42 || e.Downloads == nil || *e.Downloads != 23 {
		t.Fatalf("incorrect counts: %+v", e)
	}
	if e.License == nil || *e.License != "MIT" || e.UpdatedAt == nil || *e.UpdatedAt != "2026-01-02" {
		t.Fatalf("incorrect repository metadata: %+v", e)
	}
	if e.Author != "owner" || !e.Featured || !reflect.DeepEqual(e.Tags, ext.Tags) || !reflect.DeepEqual(e.Features, ext.Features) {
		t.Fatalf("curated metadata was not preserved: %+v", e)
	}
	if e.AddedAt == nil || *e.AddedAt != ext.AddedAt {
		t.Fatal("source addedAt was not preserved")
	}
	if e.Compatibility != ext.Compatibility {
		t.Fatal("source compatibility was not preserved")
	}
	if e.IconURL == nil || *e.IconURL != "https://images.example.org/icon.png" {
		t.Fatalf("manifest icon was not imported: %+v", e.IconURL)
	}
	wantScreenshots := []screenshot{
		{Src: "https://images.example.org/preview1.png", Alt: "Test screenshot 1"},
		{Src: "https://images.example.org/preview2.png", Alt: "Test screenshot 2"},
	}
	if !reflect.DeepEqual(e.Screenshots, wantScreenshots) {
		t.Fatalf("unexpected screenshots: %+v", e.Screenshots)
	}
}

func TestMediaURL(t *testing.T) {
	base := "https://example.org/releases/manifest.json"
	for raw, want := range map[string]string{
		"https://cdn.example.org/icon.png": "https://cdn.example.org/icon.png",
		"preview.png":                      "https://example.org/releases/preview.png",
		"../icon.png":                      "https://example.org/icon.png",
		"/images/icon.png":                 "https://example.org/images/icon.png",
		"//cdn.example.org/image.png":      "https://cdn.example.org/image.png",
		"preview 1.png":                    "https://example.org/releases/preview%201.png",
		"":                                 "",
		"  ":                               "",
		"javascript:alert(1)":              "",
		"data:image/svg+xml,<svg/>":        "",
		"http://example.org/icon.png":      "",
		"https://user:pass@example.org/a":  "",
		"https:///missing-host.png":        "",
		"https://example.org/%":            "",
	} {
		t.Run(raw, func(t *testing.T) {
			if got := mediaURL(base, raw); got != want {
				t.Fatalf("mediaURL(%q) = %q, want %q", raw, got, want)
			}
		})
	}
}

func TestReleasePagination(t *testing.T) {
	for _, failSecondPage := range []bool{false, true} {
		t.Run(fmt.Sprint(failSecondPage), func(t *testing.T) {
			pages := 0
			c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				pages++
				if r.URL.Query().Get("per_page") != "100" {
					t.Error("expected 100 releases per page")
				}
				switch r.URL.Query().Get("page") {
				case "1":
					releases := make([]map[string]any, 100)
					for i := range releases {
						releases[i] = map[string]any{"assets": []map[string]any{{"name": "app.apk", "download_count": 2}}}
					}
					json.NewEncoder(w).Encode(releases)
				case "2":
					if failSecondPage {
						http.Error(w, "rate limited", http.StatusForbidden)
						return
					}
					fmt.Fprint(w, `[{"assets":[{"name":"old.apk","download_count":7}]}]`)
				default:
					t.Error("unexpected page")
					http.NotFound(w, r)
				}
			})
			count, _, err := c.releases(context.Background(), "owner/repo")
			if pages != 2 {
				t.Fatalf("got %d pages, want 2", pages)
			}
			if failSecondPage {
				if err == nil || count != 0 {
					t.Fatalf("partial total must not be published: count=%d, err=%v", count, err)
				}
			} else if err != nil || count != 207 {
				t.Fatalf("count=%d, err=%v; want 207", count, err)
			}
		})
	}
}

func TestUnavailableMetadata(t *testing.T) {
	for _, body := range []string{"not JSON", `{"latestVersion":""}`} {
		t.Run(body, func(t *testing.T) {
			warnings := 0
			c, base := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/manifest.json" {
					fmt.Fprint(w, body)
					return
				}
				http.Error(w, "unavailable", http.StatusNotFound)
			})
			c.warn = func(string, ...any) { warnings++ }
			manifestURL := base + "/manifest.json"
			e := c.enrich(context.Background(), extension{
				Name: "Missing", Homepage: "https://github.com/owner/missing", Manifest: &manifestURL,
				Features: []string{"Fallback description"}, Tags: []string{"training"},
			})
			if warnings != 3 || e.Version != nil || e.Stars != nil || e.Downloads != nil || e.UpdatedAt != nil || e.License != nil {
				t.Fatalf("failed metadata should be unknown: warnings=%d entry=%+v", warnings, e)
			}
			if e.Description != "Fallback description" {
				t.Error("missing fallback description")
			}
			if e.IconURL != nil || len(e.Screenshots) != 0 {
				t.Fatal("unavailable manifests should not have media")
			}
		})
	}
}

func TestCuratedCompatibility(t *testing.T) {
	for _, compatibility := range []string{"both", "k2", "k3", "unknown"} {
		t.Run(compatibility, func(t *testing.T) {
			c, base := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "unavailable", http.StatusNotFound)
			})
			manifestURL := base + "/manifest.json"
			e := c.enrich(context.Background(), extension{
				Name: "Test", Homepage: "https://github.com/owner/repo", Manifest: &manifestURL,
				Features: []string{"Feature"}, Compatibility: compatibility,
			})
			if e.Compatibility != compatibility {
				t.Fatalf("failed remote metadata changed curated compatibility: got %q, want %q", e.Compatibility, compatibility)
			}
		})
	}
}

func TestZeroCountsAndNonGithub(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases") {
			fmt.Fprint(w, `[]`)
		} else {
			fmt.Fprint(w, `{"stargazers_count":0,"license":{"spdx_id":"NOASSERTION"}}`)
		}
	})
	e := c.enrich(context.Background(), extension{Homepage: "https://github.com/owner/empty", Features: []string{"Feature"}})
	if e.Stars == nil || *e.Stars != 0 || e.Downloads == nil || *e.Downloads != 0 || e.UpdatedAt != nil || e.License != nil {
		t.Fatalf("zero counts should not be unknown: %+v", e)
	}
	c.http.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Error("non-GitHub homepage with no manifest must not make requests")
		return nil, fmt.Errorf("unexpected request")
	})
	e = c.enrich(context.Background(), extension{Homepage: "https://example.org/app", Features: []string{"Feature"}})
	if e.Author != "example.org" || e.Stars != nil || e.Downloads != nil || e.Version != nil {
		t.Fatalf("non-GitHub metadata: %+v", e)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGenerateAndWrite(t *testing.T) {
	c := &metadataClient{warn: func(string, ...any) {}}
	cat := catalogue{}
	for i := range 20 {
		cat.Extensions = append(cat.Extensions, extension{
			Name: fmt.Sprint(i), Homepage: "https://example.org", Features: []string{"Feature"}, Tags: []string{"training"},
		})
	}
	entries, err := generate(context.Background(), c, cat, 4)
	if err != nil {
		t.Fatal(err)
	}
	for i, e := range entries {
		if e.Name != fmt.Sprint(i) {
			t.Fatal("parallel generation reordered entries")
		}
	}
	dir := t.TempDir()
	asset := filepath.Join(dir, "index.html")
	if err := os.WriteFile(asset, []byte("keep me"), 0o644); err != nil {
		t.Fatal(err)
	}
	builtAt := time.Date(2026, time.November, 12, 15, 4, 5, 0, time.FixedZone("UTC+2", 2*60*60))
	for range 2 {
		if err := writeCatalogue(dir, entries, builtAt); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "catalogue.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got generatedCatalogue
	if err := json.Unmarshal(data, &got); err != nil || !reflect.DeepEqual(got.Entries, entries) {
		t.Fatalf("catalogue round trip: %v", err)
	}
	if !got.BuiltAt.Equal(builtAt) || !strings.Contains(string(data), `"builtAt": "2026-11-12T13:04:05Z"`) {
		t.Fatal("build time must be serialized in UTC")
	}
	if !strings.Contains(string(data), `"downloads": null`) {
		t.Error("unknown counts must be serialized as null")
	}
	if strings.Contains(string(data), `"aiUsage"`) {
		t.Error("AI usage should no longer be exported")
	}
	if !strings.Contains(string(data), `"screenshots": []`) {
		t.Error("missing screenshots should be serialized as an empty array")
	}
	if data, _ := os.ReadFile(asset); string(data) != "keep me" {
		t.Fatal("existing site asset was overwritten")
	}
	if _, err := os.Stat(filepath.Join(dir, ".nojekyll")); err != nil {
		t.Fatal(err)
	}
	if files, _ := filepath.Glob(filepath.Join(dir, ".catalogue-*")); len(files) != 0 {
		t.Fatal("temporary output files were not cleaned up")
	}
}

func TestCancelledGeneration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := generate(ctx, &metadataClient{}, catalogue{}, 1)
	if err != context.Canceled {
		t.Fatalf("want cancellation error, got %v", err)
	}
	if _, err := generate(context.Background(), nil, catalogue{}, 0); err == nil {
		t.Fatal("zero workers must be rejected")
	}
}

func TestRequestTimeout(t *testing.T) {
	c := &metadataClient{http: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := c.getJSON(ctx, "https://example.org/manifest.json", false, new(manifest)); err == nil {
		t.Fatal("request should time out")
	}
}

func TestReadCatalogue(t *testing.T) {
	cat, err := readCatalogue("../extensions.json")
	if err != nil || len(cat.Extensions) == 0 {
		t.Fatalf("real catalogue must be readable: %v", err)
	}
	valid := `{"source":"https://example.org","extensions":[{"name":"Test","addedAt":"2024-12-21","homepage":"https://example.org","manifest":null,"features":["Feature"],"tags":["training"],"compatibility":"both","featuredInHammerheadLibrary":false}]}`
	for _, compatibility := range []string{"both", "k2", "k3", "unknown"} {
		t.Run(compatibility, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "extensions.json")
			data := strings.Replace(valid, `"compatibility":"both"`, `"compatibility":"`+compatibility+`"`, 1)
			if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
			cat, err := readCatalogue(path)
			if err != nil || cat.Extensions[0].Compatibility != compatibility {
				t.Fatalf("valid compatibility %q rejected: %v", compatibility, err)
			}
		})
	}
	for _, data := range []string{
		`{`, `{}`, `{"source":"test","extensions":[]}`, valid + `{}`,
		strings.Replace(valid, `"features":["Feature"]`, `"features":[]`, 1),
		strings.Replace(valid, `"homepage":"https://example.org"`, `"homepage":"file:///etc/passwd"`, 1),
		strings.Replace(valid, `"manifest":null`, `"manifest":"http://example.org/manifest.json"`, 1),
		strings.Replace(valid, `"name":"Test"`, `"name":"Test","unexpected":true`, 1),
		strings.Replace(valid, `"addedAt":"2024-12-21",`, ``, 1),
		strings.Replace(valid, `"2024-12-21"`, `"2026-02-30"`, 1),
		strings.Replace(valid, `"2024-12-21"`, `"2026-2-3"`, 1),
		strings.Replace(valid, `"2024-12-21"`, `"2026-02-3"`, 1),
		strings.Replace(valid, `"compatibility":"both",`, ``, 1),
		strings.Replace(valid, `"compatibility":"both"`, `"compatibility":"Karoo 2"`, 1),
		strings.Replace(valid, `"compatibility":"both"`, `"compatibility":null`, 1),
	} {
		t.Run(data, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "extensions.json")
			if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := readCatalogue(path); err == nil {
				t.Fatal("invalid catalogue was accepted")
			}
		})
	}
}

func TestGetJSONRejectsTrailingGarbage(t *testing.T) {
	c := &metadataClient{http: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{} garbage`))}, nil
	})}}
	if err := c.getJSON(context.Background(), "https://example.org/manifest.json", false, new(manifest)); err == nil {
		t.Fatal("invalid JSON response accepted")
	}
}

func TestRateLimitSkipsAPIRequestsButNotManifests(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusTooManyRequests} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			requests := 0
			c, base := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.URL.Path == "/manifest.json" {
					fmt.Fprint(w, `{"latestVersion":"1.0"}`)
					return
				}
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, "rate limited", status)
			})
			for range 2 {
				if err := c.getJSON(context.Background(), base+"/repos/owner/repo", true, new(repository)); err == nil {
					t.Fatal("rate-limited API call should fail")
				}
			}
			if requests != 1 {
				t.Fatalf("rate-limited client made %d API requests, want 1", requests)
			}
			var m manifest
			if err := c.getJSON(context.Background(), base+"/manifest.json", false, &m); err != nil || m.LatestVersion != "1.0" {
				t.Fatalf("manifest requests should still work: %v", err)
			}
		})
	}
}
