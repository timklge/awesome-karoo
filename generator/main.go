// Command generator builds the static site. All remote metadata is resolved
// here, never by visitors' browsers.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type extension struct {
	Name          string   `json:"name"`
	AddedAt       string   `json:"addedAt"`
	Homepage      string   `json:"homepage"`
	Manifest      *string  `json:"manifest"`
	Features      []string `json:"features"`
	Tags          []string `json:"tags"`
	Compatibility string   `json:"compatibility"`
	Featured      bool     `json:"featuredInHammerheadLibrary"`
}

type catalogue struct {
	Schema     string      `json:"$schema"`
	Source     string      `json:"source"`
	Extensions []extension `json:"extensions"`
}

// Keep the JSON shape consumed by static/index.html. Nil values mean unknown,
// which is different from a repository with zero stars or downloads.
type entry struct {
	Name          string       `json:"name"`
	URL           string       `json:"url"`
	Manifest      *string      `json:"manifest"`
	Description   string       `json:"description"`
	Features      []string     `json:"features"`
	Tags          []string     `json:"tags"`
	Compatibility string       `json:"compatibility"`
	License       *string      `json:"license"`
	Featured      bool         `json:"featured"`
	Version       *string      `json:"version"`
	AddedAt       *string      `json:"addedAt"`
	UpdatedAt     *string      `json:"updatedAt"`
	IconURL       *string      `json:"iconUrl"`
	Screenshots   []screenshot `json:"screenshots"`
	Author        string       `json:"author"`
	Stars         *int64       `json:"stars"`
	Downloads     *int64       `json:"downloads"`
}

type generatedCatalogue struct {
	BuiltAt time.Time `json:"builtAt"`
	Entries []entry   `json:"entries"`
}

type manifest struct {
	LatestVersion  string   `json:"latestVersion"`
	Description    string   `json:"description"`
	IconURL        string   `json:"iconUrl"`
	ScreenshotURLs []string `json:"screenshotUrls"`
}

type screenshot struct {
	Src string `json:"src"`
	Alt string `json:"alt"`
}

// Resolve manifest-relative image paths, accepting only HTTPS URLs without credentials.
func mediaURL(manifestURL, raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	base, err := url.Parse(manifestURL)
	if err != nil {
		return ""
	}
	reference, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	u := base.ResolveReference(reference)
	if u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return ""
	}
	return u.String()
}

type repository struct {
	Stars   *int64 `json:"stargazers_count"`
	License *struct {
		SPDX string `json:"spdx_id"`
	} `json:"license"`
}

type release struct {
	Draft       bool       `json:"draft"`
	Prerelease  bool       `json:"prerelease"`
	PublishedAt *time.Time `json:"published_at"`
	Assets      []struct {
		Name      string `json:"name"`
		Downloads int64  `json:"download_count"`
	} `json:"assets"`
}

type metadataClient struct {
	http           *http.Client
	apiBase        string
	token          string
	warn           func(string, ...any)
	apiRateLimited atomic.Bool
}

var repoPart = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

func githubRepo(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Host, "github.com") || u.User != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	parts[1] = strings.TrimSuffix(parts[1], ".git")
	for _, part := range parts[:2] {
		if !repoPart.MatchString(part) || part == "." || part == ".." {
			return ""
		}
	}
	return parts[0] + "/" + parts[1]
}

func (c *metadataClient) getJSON(ctx context.Context, address string, github bool, target any) error {
	if github && c.apiRateLimited.Load() {
		return fmt.Errorf("GitHub API rate limit exhausted; skipping remaining API requests")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "awesome-karoo-generator")
	req.Header.Set("Accept", "application/json")
	if github {
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if github && (resp.StatusCode == http.StatusTooManyRequests ||
			(resp.StatusCode == http.StatusForbidden && (resp.Header.Get("X-RateLimit-Remaining") == "0" || resp.Header.Get("Retry-After") != ""))) {
			c.apiRateLimited.Store(true)
		}
		if remaining := resp.Header.Get("X-RateLimit-Remaining"); github && remaining != "" {
			return fmt.Errorf("HTTP %d from %s (GitHub rate limit remaining: %s)", resp.StatusCode, address, remaining)
		}
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, address)
	}
	// Bound memory usage even if a remote manifest is unexpectedly large.
	const maxJSONSize = 32 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxJSONSize+1))
	if err != nil {
		return err
	}
	if len(body) > maxJSONSize {
		return fmt.Errorf("JSON response from %s exceeds %d bytes", address, maxJSONSize)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode %s: %w", address, err)
	}
	return nil
}

func (c *metadataClient) releases(ctx context.Context, repo string) (int64, *string, error) {
	var downloads int64
	var latest time.Time
	for page := 1; ; page++ {
		var releases []release
		address := fmt.Sprintf("%s/repos/%s/releases?per_page=100&page=%d", c.apiBase, repo, page)
		if err := c.getJSON(ctx, address, true, &releases); err != nil {
			// Never publish a partial total if a later page fails.
			return 0, nil, err
		}
		for _, r := range releases {
			if r.Draft {
				continue
			}
			// Include prerelease APKs in the lifetime download total, but use
			// stable releases for the displayed update date.
			for _, asset := range r.Assets {
				if strings.HasSuffix(strings.ToLower(asset.Name), ".apk") {
					downloads += asset.Downloads
				}
			}
			if !r.Prerelease && r.PublishedAt != nil && r.PublishedAt.After(latest) {
				latest = *r.PublishedAt
			}
		}
		if len(releases) < 100 {
			break
		}
	}
	var updated *string
	if !latest.IsZero() {
		date := latest.UTC().Format(time.DateOnly)
		updated = &date
	}
	return downloads, updated, nil
}

func (c *metadataClient) enrich(ctx context.Context, ext extension) entry {
	e := entry{
		Name: ext.Name, URL: ext.Homepage, Manifest: ext.Manifest,
		Description: ext.Features[0], Features: ext.Features, Tags: ext.Tags,
		Featured: ext.Featured, Compatibility: ext.Compatibility,
		Screenshots: []screenshot{},
	}
	if ext.AddedAt != "" {
		e.AddedAt = &ext.AddedAt
	}
	if u, err := url.Parse(ext.Homepage); err == nil {
		e.Author = u.Hostname()
	}
	if ext.Manifest != nil {
		var m manifest
		if err := c.getJSON(ctx, *ext.Manifest, false, &m); err != nil {
			c.warn("%s: manifest: %v", ext.Name, err)
		} else {
			if version := strings.TrimSpace(m.LatestVersion); version != "" {
				e.Version = &version
			} else {
				c.warn("%s: manifest has no latestVersion", ext.Name)
			}
			if strings.TrimSpace(m.Description) != "" {
				e.Description = m.Description
			}
			if icon := mediaURL(*ext.Manifest, m.IconURL); icon != "" {
				e.IconURL = &icon
			}
			seen := make(map[string]bool)
			for i, raw := range m.ScreenshotURLs {
				if src := mediaURL(*ext.Manifest, raw); src != "" && !seen[src] {
					seen[src] = true
					e.Screenshots = append(e.Screenshots, screenshot{
						Src: src, Alt: fmt.Sprintf("%s screenshot %d", ext.Name, i+1),
					})
				}
			}
		}
	}
	repo := githubRepo(ext.Homepage)
	if repo == "" {
		return e
	}
	e.Author = strings.SplitN(repo, "/", 2)[0]
	var r repository
	if err := c.getJSON(ctx, c.apiBase+"/repos/"+repo, true, &r); err != nil {
		c.warn("%s: repository: %v", ext.Name, err)
	} else {
		e.Stars = r.Stars
		if r.License != nil && r.License.SPDX != "" && r.License.SPDX != "NOASSERTION" {
			e.License = &r.License.SPDX
		}
	}
	if downloads, updated, err := c.releases(ctx, repo); err != nil {
		c.warn("%s: releases: %v", ext.Name, err)
	} else {
		e.Downloads, e.UpdatedAt = &downloads, updated
	}
	return e
}

func readCatalogue(path string) (catalogue, error) {
	var cat catalogue
	data, err := os.ReadFile(path)
	if err != nil {
		return cat, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cat); err != nil {
		return cat, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return cat, fmt.Errorf("expected a single JSON document")
	}
	if cat.Source == "" || len(cat.Extensions) == 0 {
		return cat, fmt.Errorf("source and a nonempty extensions array are required")
	}
	for i, ext := range cat.Extensions {
		switch ext.Compatibility {
		case "both", "k2", "k3", "unknown":
		default:
			return cat, fmt.Errorf("%s: compatibility must be both, k2, k3 or unknown", ext.Name)
		}
		if date, err := time.Parse(time.DateOnly, ext.AddedAt); err != nil || date.Format(time.DateOnly) != ext.AddedAt {
			return cat, fmt.Errorf("%s: addedAt must be a valid YYYY-MM-DD date", ext.Name)
		}
		if strings.TrimSpace(ext.Name) == "" || len(ext.Features) == 0 || len(ext.Tags) == 0 {
			return cat, fmt.Errorf("extension %d: name, features and tags are required", i+1)
		}
		addresses := []string{ext.Homepage}
		if ext.Manifest != nil {
			addresses = append(addresses, *ext.Manifest)
		}
		for _, address := range addresses {
			u, err := url.Parse(address)
			if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
				return cat, fmt.Errorf("%s: invalid HTTPS URL %q", ext.Name, address)
			}
		}
	}
	return cat, nil
}

func generate(ctx context.Context, c *metadataClient, cat catalogue, workers int) ([]entry, error) {
	if workers < 1 {
		return nil, fmt.Errorf("workers must be positive")
	}
	entries := make([]entry, len(cat.Extensions))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for index := range jobs {
				entries[index] = c.enrich(ctx, cat.Extensions[index])
			}
		})
	}
	for i := range cat.Extensions {
		select {
		case jobs <- i:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		}
	}
	close(jobs)
	wg.Wait()
	return entries, ctx.Err()
}

const cataloguePlaceholder = "<!-- CATALOGUE_JSON -->"

func writeSite(source, dir string, entries []entry, builtAt time.Time) error {
	source, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	if rel, err := filepath.Rel(source, dir); err != nil {
		return err
	} else if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output directory must be outside the static source directory")
	}
	index, err := os.ReadFile(filepath.Join(source, "index.html"))
	if err != nil {
		return err
	}
	if bytes.Count(index, []byte(cataloguePlaceholder)) != 1 {
		return fmt.Errorf("index.html must contain exactly one %s placeholder", cataloguePlaceholder)
	}
	// Keep encoding/json's HTML escaping so remote metadata cannot close the
	// application/json script element or inject HTML into the generated page.
	data, err := json.MarshalIndent(generatedCatalogue{BuiltAt: builtAt.UTC(), Entries: entries}, "", "  ")
	if err != nil {
		return err
	}
	index = bytes.Replace(index, []byte(cataloguePlaceholder), data, 1)
	if err := filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dir, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		// The index and catalogue are generated, not copied from old builds.
		if rel == "index.html" || rel == "catalogue.json" {
			return nil
		}
		asset, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, asset, 0o644)
	}); err != nil {
		return err
	}
	if err := writeCatalogue(dir, data); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "index.html"), index, 0o644)
}

func writeCatalogue(dir string, data []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Replace the catalogue atomically without touching the site's other assets.
	f, err := os.CreateTemp(dir, ".catalogue-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(append(data, '\n')); err != nil {
		f.Close()
		return err
	}
	if err := f.Chmod(0o644); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), filepath.Join(dir, "catalogue.json")); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".nojekyll"), nil, 0o644)
}

func run() error {
	input := flag.String("input", "../extensions.json", "path to the extension catalogue")
	source := flag.String("static", "../static", "directory containing the site's source assets")
	output := flag.String("output", "output", "directory to write the generated site into")
	workers := flag.Int("workers", 4, "number of extensions to fetch concurrently")
	timeout := flag.Duration("timeout", 30*time.Second, "timeout per HTTP request")
	flag.Parse()
	if *workers < 1 || *timeout <= 0 {
		return fmt.Errorf("workers and timeout must be positive")
	}
	cat, err := readCatalogue(*input)
	if err != nil {
		return fmt.Errorf("read catalogue: %w", err)
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		log.Print("warning: no GITHUB_TOKEN or GH_TOKEN; GitHub's anonymous API limit is usually too low for the full catalogue")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	c := &metadataClient{
		http: &http.Client{Timeout: *timeout}, apiBase: "https://api.github.com",
		token: token, warn: log.Printf,
	}
	entries, err := generate(ctx, c, cat, *workers)
	if err != nil {
		return err
	}
	if err := writeSite(*source, *output, entries, time.Now()); err != nil {
		return fmt.Errorf("write site: %w", err)
	}
	log.Printf("generated site in %s with %d extensions", *output, len(entries))
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
