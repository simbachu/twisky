package feed

// NewPostsSince returns posts from the start of posts up to (but not including)
// the post matching sinceID. Returns nil when sinceID is empty. When sinceID is
// not found in posts, returns the whole page (treat as all new).
func NewPostsSince(posts []PostView, sinceID string) []PostView {
	if sinceID == "" {
		return nil
	}
	for index, post := range posts {
		if post.ID == sinceID {
			if index == 0 {
				return nil
			}
			return posts[:index]
		}
	}
	return posts
}
