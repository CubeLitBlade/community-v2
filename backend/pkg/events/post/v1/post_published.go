package v1

// PostPublished is the event payload emitted when a new post is published.
type PostPublished struct {
	PostID   string `json:"postId"`
	AuthorID string `json:"authorId"`
	Title    string `json:"title,omitempty"`
}
