package tiengtrungonline_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

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

const fakeSearch = `[
  {"id":34749,"title":"4 App giup ban hoc cach phat am va binh am Pinyin tieng Trung","url":"https://tiengtrungonline.com/app-hoc.html","type":"post","subtype":"post"},
  {"id":16723,"title":"Chinesepod Pinyin Chart: Phan mem doc phat am","url":"https://tiengtrungonline.com/chinesepod.html","type":"post","subtype":"post"},
  {"id":999,"title":"Ignored page","url":"https://tiengtrungonline.com/page","type":"page","subtype":"page"}
]`

const fakeLesson = `[{
  "id": 70751,
  "slug": "bai-12-toi-muon-di-bo-phieu",
  "date": "2026-05-17T15:41:37",
  "link": "https://tiengtrungonline.com/bai-12-toi-muon-di-bo-phieu.html",
  "title": {"rendered": "Bai 12: Toi muon di bo phieu"},
  "excerpt": {"rendered": "<p>Trong bai hoc tiep theo cua Giao trinh...</p>"},
  "content": {"rendered": "<p>Ngu phap</p><p>Some content here with more words and text.</p>"}
}]`

const fakeSiteRoot = `{
  "name": "Trung tam tieng Trung Chinese",
  "description": "Website Chinese la noi giao vien chia se tai lieu",
  "url": "https://tiengtrungonline.com"
}`

func newTestClient(ts *httptest.Server) *tiengtrungonline.Client {
	cfg := tiengtrungonline.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return tiengtrungonline.NewClient(cfg)
}

func TestPosts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakePosts)
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
		_, _ = fmt.Fprint(w, fakeCategories)
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

func TestSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeSearch)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	results, err := c.Search(context.Background(), "pinyin", 20)
	if err != nil {
		t.Fatal(err)
	}
	// Only "post" type results (not "page")
	if len(results) != 2 {
		t.Fatalf("want 2 post results, got %d", len(results))
	}
	if results[0].ID != 34749 {
		t.Errorf("ID = %d", results[0].ID)
	}
	if results[0].Rank != 1 {
		t.Errorf("Rank = %d", results[0].Rank)
	}
	if results[1].Rank != 2 {
		t.Errorf("second Rank = %d", results[1].Rank)
	}
}

func TestLesson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeLesson)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	l, err := c.Lesson(context.Background(), "bai-12-toi-muon-di-bo-phieu")
	if err != nil {
		t.Fatal(err)
	}
	if l.ID != 70751 {
		t.Errorf("ID = %d", l.ID)
	}
	if l.Date != "2026-05-17" {
		t.Errorf("Date = %q", l.Date)
	}
	if l.Title != "Bai 12: Toi muon di bo phieu" {
		t.Errorf("Title = %q", l.Title)
	}
	if l.ContentLength == 0 {
		t.Error("ContentLength should be > 0")
	}
	if l.Excerpt == "" {
		t.Error("Excerpt should not be empty")
	}
}

func TestSiteInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-WP-Total", "3641")
		switch r.URL.Path {
		case "/wp-json/":
			_, _ = fmt.Fprint(w, fakeSiteRoot)
		default:
			_, _ = fmt.Fprint(w, "[]")
		}
	}))
	defer ts.Close()

	c := newTestClient(ts)
	info, err := c.SiteInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "Trung tam tieng Trung Chinese" {
		t.Errorf("Name = %q", info.Name)
	}
	if info.Posts != 3641 {
		t.Errorf("Posts = %d, want 3641", info.Posts)
	}
	if info.Categories != 3641 {
		t.Errorf("Categories = %d, want 3641", info.Categories)
	}
}

func TestPostsWithTotal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-WP-Total", "42")
		_, _ = fmt.Fprint(w, fakePosts)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	r, err := c.PostsWithTotal(context.Background(), 10, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 42 {
		t.Errorf("Total = %d, want 42", r.Total)
	}
	if len(r.Posts) != 2 {
		t.Errorf("Posts count = %d, want 2", len(r.Posts))
	}
}

func TestRetryOn429(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = fmt.Fprint(w, "[]")
	}))
	defer ts.Close()

	cfg := tiengtrungonline.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := tiengtrungonline.NewClient(cfg)

	// Use Search which doesn't trigger the category cache fetch.
	start := time.Now()
	_, err := c.Search(context.Background(), "test", 10)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("hits = %d, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("no backoff between retries")
	}
}

func TestNonRetriableOn404(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	cfg := tiengtrungonline.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := tiengtrungonline.NewClient(cfg)

	// Use Search which doesn't trigger category cache.
	_, err := c.Search(context.Background(), "test", 10)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if hits != 1 {
		t.Errorf("hits = %d on 404, want exactly 1", hits)
	}
}

func TestPace(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "[]")
	}))
	defer ts.Close()

	cfg := tiengtrungonline.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 100 * time.Millisecond
	cfg.Retries = 0
	c := tiengtrungonline.NewClient(cfg)

	// Use Search for both calls to avoid category cache side effects.
	start := time.Now()
	_, _ = c.Search(context.Background(), "test", 1)
	_, _ = c.Search(context.Background(), "test2", 1)
	if time.Since(start) < 90*time.Millisecond {
		t.Error("second request too fast: pace() not working")
	}
}

func TestExportPagination(t *testing.T) {
	page1 := `[{"id":1,"date":"2026-01-01T00:00:00","slug":"p1","link":"https://x.com/1","title":{"rendered":"Post 1"},"categories":[]}]`
	page2 := `[{"id":2,"date":"2026-01-02T00:00:00","slug":"p2","link":"https://x.com/2","title":{"rendered":"Post 2"},"categories":[]}]`

	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		pageParam := r.URL.Query().Get("page")
		p, _ := strconv.Atoi(pageParam)
		switch p {
		case 1:
			w.Header().Set("X-WP-Total", "2")
			_, _ = fmt.Fprint(w, page1)
		case 2:
			w.Header().Set("X-WP-Total", "2")
			_, _ = fmt.Fprint(w, page2)
		default:
			w.Header().Set("X-WP-Total", "2")
			_, _ = fmt.Fprint(w, "[]")
		}
	}))
	defer ts.Close()

	c := newTestClient(ts)

	var all []*tiengtrungonline.Post
	for page := 1; ; page++ {
		r, err := c.PostsWithTotal(context.Background(), 1, page, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Posts) == 0 {
			break
		}
		all = append(all, r.Posts...)
		if len(r.Posts) < 1 {
			break
		}
	}

	if len(all) != 2 {
		t.Errorf("got %d posts, want 2", len(all))
	}
	if all[0].Title != "Post 1" {
		t.Errorf("first post title = %q", all[0].Title)
	}
	if all[1].Title != "Post 2" {
		t.Errorf("second post title = %q", all[1].Title)
	}
}
