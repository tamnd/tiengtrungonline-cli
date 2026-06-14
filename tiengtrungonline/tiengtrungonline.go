// Package tiengtrungonline is the library behind the tiengtrungonline CLI.
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
	"time"
)

const DefaultUserAgent = "tiengtrungonline-cli/dev (+https://github.com/tamnd/tiengtrungonline-cli)"

type Config struct {
	BaseURL   string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://tiengtrungonline.com",
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
		UserAgent: DefaultUserAgent,
	}
}

type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

var tagRE = regexp.MustCompile(`<[^>]+>`)

// internal API types

type apiPost struct {
	ID    int    `json:"id"`
	Date  string `json:"date"`
	Slug  string `json:"slug"`
	Link  string `json:"link"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Categories []int `json:"categories"`
}

type apiCategory struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Count int    `json:"count"`
}

// Posts fetches posts from the WordPress REST API.
func (c *Client) Posts(ctx context.Context, perPage, page, categoryID int) ([]*Post, error) {
	url := fmt.Sprintf("%s/wp-json/wp/v2/posts?per_page=%d&page=%d", c.cfg.BaseURL, perPage, page)
	if categoryID != 0 {
		url += fmt.Sprintf("&categories=%d", categoryID)
	}
	body, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var raw []apiPost
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode posts: %w", err)
	}
	rank := (page-1)*perPage + 1
	var posts []*Post
	for _, ap := range raw {
		date := ap.Date
		if len(date) >= 10 {
			date = date[:10]
		}
		title := html.UnescapeString(ap.Title.Rendered)
		title = tagRE.ReplaceAllString(title, "")
		posts = append(posts, &Post{
			Rank:     rank,
			Date:     date,
			Category: "",
			Title:    title,
			URL:      ap.Link,
		})
		rank++
	}
	return posts, nil
}

// Categories fetches all categories from the WordPress REST API.
func (c *Client) Categories(ctx context.Context) ([]*Category, error) {
	url := fmt.Sprintf("%s/wp-json/wp/v2/categories?per_page=100", c.cfg.BaseURL)
	body, err := c.get(ctx, url)
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

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}
