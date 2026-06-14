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
