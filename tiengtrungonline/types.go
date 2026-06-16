package tiengtrungonline

// Post is one article from Tieng Trung Online.
type Post struct {
	Rank     int    `json:"rank"     csv:"rank"     tsv:"rank"`
	Date     string `json:"date"     csv:"date"     tsv:"date"`
	Category string `json:"category" csv:"category" tsv:"category"`
	Title    string `json:"title"    csv:"title"    tsv:"title"`
	URL      string `json:"url"      csv:"url"      tsv:"url"`
}

// Category is one category from Tieng Trung Online.
type Category struct {
	Rank  int    `json:"rank"  csv:"rank"  tsv:"rank"`
	ID    int    `json:"id"    csv:"id"    tsv:"id"`
	Slug  string `json:"slug"  csv:"slug"  tsv:"slug"`
	Name  string `json:"name"  csv:"name"  tsv:"name"`
	Count int    `json:"count" csv:"count" tsv:"count"`
}

// SearchResult is one item from the WP search API.
type SearchResult struct {
	Rank  int    `json:"rank"  csv:"rank"  tsv:"rank"`
	ID    int    `json:"id"    csv:"id"    tsv:"id"`
	Title string `json:"title" csv:"title" tsv:"title"`
	URL   string `json:"url"   csv:"url"   tsv:"url"`
}

// Lesson is a full post record with content details.
type Lesson struct {
	ID            int    `json:"id"             csv:"id"             tsv:"id"`
	Slug          string `json:"slug"           csv:"slug"           tsv:"slug"`
	Date          string `json:"date"           csv:"date"           tsv:"date"`
	Title         string `json:"title"          csv:"title"          tsv:"title"`
	URL           string `json:"url"            csv:"url"            tsv:"url"`
	Excerpt       string `json:"excerpt"        csv:"excerpt"        tsv:"excerpt"`
	ContentLength int    `json:"content_length" csv:"content_length" tsv:"content_length"`
}

// SiteInfo is summary statistics for the site.
type SiteInfo struct {
	Name        string `json:"name"        csv:"name"        tsv:"name"`
	URL         string `json:"url"         csv:"url"         tsv:"url"`
	Posts       int    `json:"posts"       csv:"posts"       tsv:"posts"`
	Categories  int    `json:"categories"  csv:"categories"  tsv:"categories"`
	Description string `json:"description" csv:"description" tsv:"description"`
}

// PostsResult bundles paginated posts with the total count from X-WP-Total.
type PostsResult struct {
	Posts []*Post
	Total int
}
