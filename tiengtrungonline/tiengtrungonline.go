// Package tiengtrungonline is the library behind the tto command line.
package tiengtrungonline

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const DefaultUserAgent = "tto/dev (+https://github.com/tamnd/tiengtrungonline-cli)"

// Config holds the client configuration.
type Config struct {
	BaseURL   string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://tiengtrungonline.com",
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
		UserAgent: DefaultUserAgent,
	}
}

// Client talks to tiengtrungonline.com via the WordPress REST API.
type Client struct {
	cfg      Config
	http     *http.Client
	mu       sync.Mutex
	last     time.Time
	catOnce  sync.Once
	catCache map[string]int // slug -> ID
}

// NewClient returns a new Client.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

var tagRE = regexp.MustCompile(`<[^>]+>`)

// --- internal API wire types ---

type apiPost struct {
	ID    int    `json:"id"`
	Date  string `json:"date"`
	Slug  string `json:"slug"`
	Link  string `json:"link"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Excerpt struct {
		Rendered string `json:"rendered"`
	} `json:"excerpt"`
	Content struct {
		Rendered string `json:"rendered"`
	} `json:"content"`
	Categories []int `json:"categories"`
}

type apiCategory struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Count int    `json:"count"`
}

type apiSearchResult struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

type wpApiRoot struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// --- public methods ---

// PostsWithTotal fetches a page of posts and the total count from X-WP-Total.
func (c *Client) PostsWithTotal(ctx context.Context, perPage, page, catID int) (*PostsResult, error) {
	u := fmt.Sprintf("%s/wp-json/wp/v2/posts?per_page=%d&page=%d&_fields=id,slug,date,link,title,categories",
		c.cfg.BaseURL, perPage, page)
	if catID != 0 {
		u += fmt.Sprintf("&categories=%d", catID)
	}
	body, total, err := c.getWithTotal(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw []apiPost
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode posts: %w", err)
	}
	offset := (page-1)*perPage + 1
	catMap, _ := c.categoryMap(ctx)
	var posts []*Post
	for i, ap := range raw {
		date := ap.Date
		if len(date) >= 10 {
			date = date[:10]
		}
		title := cleanHTML(ap.Title.Rendered)
		catSlug := ""
		if len(ap.Categories) > 0 {
			catSlug = catMap[ap.Categories[0]]
		}
		posts = append(posts, &Post{
			Rank:     offset + i,
			Date:     date,
			Category: catSlug,
			Title:    title,
			URL:      ap.Link,
		})
	}
	return &PostsResult{Posts: posts, Total: total}, nil
}

// Posts fetches posts from the WordPress REST API.
func (c *Client) Posts(ctx context.Context, perPage, page, categoryID int) ([]*Post, error) {
	r, err := c.PostsWithTotal(ctx, perPage, page, categoryID)
	if err != nil {
		return nil, err
	}
	return r.Posts, nil
}

// Categories fetches all categories from the WordPress REST API.
func (c *Client) Categories(ctx context.Context) ([]*Category, error) {
	u := fmt.Sprintf("%s/wp-json/wp/v2/categories?per_page=100", c.cfg.BaseURL)
	body, _, err := c.getWithTotal(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw []apiCategory
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode categories: %w", err)
	}
	var cats []*Category
	for _, ac := range raw {
		if ac.Count == 0 {
			continue
		}
		cats = append(cats, &Category{
			ID:    ac.ID,
			Slug:  ac.Slug,
			Name:  ac.Name,
			Count: ac.Count,
		})
	}
	sort.Slice(cats, func(i, j int) bool {
		return cats[i].Count > cats[j].Count
	})
	for i, cat := range cats {
		cat.Rank = i + 1
	}
	return cats, nil
}

// Search searches posts by title using the WP search API.
func (c *Client) Search(ctx context.Context, query string, perPage int) ([]*SearchResult, error) {
	u := fmt.Sprintf("%s/wp-json/wp/v2/search?search=%s&type=post&per_page=%d",
		c.cfg.BaseURL, queryEscape(query), perPage)
	body, _, err := c.getWithTotal(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw []apiSearchResult
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode search: %w", err)
	}
	var out []*SearchResult
	for i, r := range raw {
		if r.Type != "post" {
			continue
		}
		out = append(out, &SearchResult{
			Rank:  i + 1,
			ID:    r.ID,
			Title: r.Title,
			URL:   r.URL,
		})
	}
	return out, nil
}

// Lesson fetches a single post by slug.
func (c *Client) Lesson(ctx context.Context, slug string) (*Lesson, error) {
	slug = strings.TrimSuffix(slug, ".html")
	slug = strings.TrimPrefix(slug, "/")
	u := fmt.Sprintf("%s/wp-json/wp/v2/posts?slug=%s&_fields=id,slug,date,link,title,excerpt,content",
		c.cfg.BaseURL, queryEscape(slug))
	body, _, err := c.getWithTotal(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw []apiPost
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode lesson: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("post not found: %s", slug)
	}
	ap := raw[0]
	date := ap.Date
	if len(date) >= 10 {
		date = date[:10]
	}
	excerpt := cleanHTMLTrunc(ap.Excerpt.Rendered, 200)
	return &Lesson{
		ID:            ap.ID,
		Slug:          ap.Slug,
		Date:          date,
		Title:         cleanHTML(ap.Title.Rendered),
		URL:           ap.Link,
		Excerpt:       excerpt,
		ContentLength: wordCount(ap.Content.Rendered),
	}, nil
}

// SiteInfo fetches site-level statistics.
func (c *Client) SiteInfo(ctx context.Context) (*SiteInfo, error) {
	// WP API root for name/description.
	rootBody, _, err := c.getWithTotal(ctx, c.cfg.BaseURL+"/wp-json/")
	if err != nil {
		return nil, err
	}
	var root wpApiRoot
	if err := json.Unmarshal(rootBody, &root); err != nil {
		return nil, fmt.Errorf("decode wp root: %w", err)
	}

	// Post count from X-WP-Total.
	_, postTotal, err := c.getWithTotal(ctx, c.cfg.BaseURL+"/wp-json/wp/v2/posts?per_page=1")
	if err != nil {
		return nil, err
	}

	// Category count from X-WP-Total.
	_, catTotal, err := c.getWithTotal(ctx, c.cfg.BaseURL+"/wp-json/wp/v2/categories?per_page=1")
	if err != nil {
		return nil, err
	}

	siteURL := root.URL
	if siteURL == "" {
		siteURL = c.cfg.BaseURL
	}
	return &SiteInfo{
		Name:        root.Name,
		URL:         siteURL,
		Posts:       postTotal,
		Categories:  catTotal,
		Description: root.Description,
	}, nil
}

// CategoryID resolves a category slug to its ID. Results are cached.
func (c *Client) CategoryID(ctx context.Context, slug string) (int, error) {
	c.catOnce.Do(func() {
		cats, err := c.Categories(ctx)
		if err != nil {
			return
		}
		c.catCache = make(map[string]int, len(cats))
		for _, cat := range cats {
			c.catCache[cat.Slug] = cat.ID
		}
	})
	if c.catCache == nil {
		return 0, fmt.Errorf("categories unavailable")
	}
	id, ok := c.catCache[slug]
	if !ok {
		return 0, fmt.Errorf("category %q not found", slug)
	}
	return id, nil
}

// --- internal methods ---

func (c *Client) categoryMap(ctx context.Context) (map[int]string, error) {
	c.catOnce.Do(func() {
		cats, err := c.Categories(ctx)
		if err != nil {
			return
		}
		c.catCache = make(map[string]int, len(cats))
		for _, cat := range cats {
			c.catCache[cat.Slug] = cat.ID
		}
	})
	if c.catCache == nil {
		return map[int]string{}, nil
	}
	out := make(map[int]string, len(c.catCache))
	for slug, id := range c.catCache {
		out[id] = slug
	}
	return out, nil
}

func (c *Client) getWithTotal(ctx context.Context, u string) ([]byte, int, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, total, retry, err := c.doWithTotal(ctx, u)
		if err == nil {
			return body, total, nil
		}
		lastErr = err
		if !retry {
			return nil, 0, err
		}
	}
	return nil, 0, fmt.Errorf("get %s: %w", u, lastErr)
}

func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	b, _, err := c.getWithTotal(ctx, u)
	return b, err
}

func (c *Client) doWithTotal(ctx context.Context, u string) ([]byte, int, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, 0, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	total := 0
	if h := resp.Header.Get("X-WP-Total"); h != "" {
		total, _ = strconv.Atoi(h)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, true, err
	}
	return b, total, false, nil
}

func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		return 5 * time.Second
	}
	return d
}

// --- helpers ---

func cleanHTML(s string) string {
	s = html.UnescapeString(s)
	s = tagRE.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func cleanHTMLTrunc(s string, maxLen int) string {
	s = cleanHTML(s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}

func wordCount(htmlStr string) int {
	plain := tagRE.ReplaceAllString(htmlStr, " ")
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return 0
	}
	return len(strings.Fields(plain))
}

func queryEscape(s string) string {
	// Simple percent-encoding for URL query values.
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.', r == '~':
			b.WriteRune(r)
		default:
			for _, by := range []byte(string(r)) {
				fmt.Fprintf(&b, "%%%02X", by)
			}
		}
	}
	return b.String()
}
