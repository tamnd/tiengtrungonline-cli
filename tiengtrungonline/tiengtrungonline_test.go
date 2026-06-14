package tiengtrungonline_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/tiengtrungonline-cli/tiengtrungonline"
)

const fakePosts = `[
  {"id":1,"date":"2026-05-17T15:41:37","slug":"bai-1","link":"https://example.com/bai-1","title":{"rendered":"B&#224;i 1: Xin ch&#224;o"},"categories":[507]},
  {"id":2,"date":"2026-04-01T10:00:00","slug":"bai-2","link":"https://example.com/bai-2","title":{"rendered":"B&#224;i 2: C&#225;m &#417;n"},"categories":[509]}
]`

const fakeCategories = `[
  {"id":507,"name":"Đương đại 3","slug":"duong-dai-3","count":36},
  {"id":509,"name":"Đương đại 1","slug":"duong-dai-1","count":45},
  {"id":1,"name":"Empty","slug":"empty","count":0}
]`

func newTestClient(ts *httptest.Server) *tiengtrungonline.Client {
	cfg := tiengtrungonline.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return tiengtrungonline.NewClient(cfg)
}

func TestPosts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakePosts)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	posts, err := c.Posts(context.Background(), 10, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 {
		t.Fatalf("want 2, got %d", len(posts))
	}
	if posts[0].Date != "2026-05-17" {
		t.Errorf("Date = %q", posts[0].Date)
	}
	if posts[0].Title != "Bài 1: Xin chào" {
		t.Errorf("Title = %q", posts[0].Title)
	}
	if posts[0].Rank != 1 {
		t.Errorf("Rank = %d", posts[0].Rank)
	}
}

func TestCategories(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeCategories)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	cats, err := c.Categories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// count==0 category is skipped
	if len(cats) != 2 {
		t.Fatalf("want 2, got %d", len(cats))
	}
	// Sorted by Count desc
	if cats[0].Count < cats[1].Count {
		t.Errorf("not sorted by count desc: %d < %d", cats[0].Count, cats[1].Count)
	}
}
