package http

import (
	"log/slog"
	"net/http"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/tag"
)

func (s *Server) enrichFeedLikes(r *http.Request, feed *feedquery.FeedView) {
	if feed == nil || len(feed.Posts) == 0 {
		return
	}
	client := s.sessionClient(r)
	if client == nil {
		return
	}
	uris := feedquery.CollectPostURIs(feed.Posts)
	posts, err := client.GetPosts(r.Context(), uris)
	if err != nil {
		slog.Warn("viewer like enrichment failed", "err", err)
		return
	}
	feedquery.ApplyViewerLikes(feed.Posts, posts)
}

func (s *Server) enrichPageLikes(r *http.Request, view *feedquery.PostPageView) {
	if view == nil {
		return
	}
	client := s.sessionClient(r)
	if client == nil {
		return
	}
	uris := feedquery.CollectPagePostURIs(*view)
	posts, err := client.GetPosts(r.Context(), uris)
	if err != nil {
		slog.Warn("viewer like enrichment failed", "err", err)
		return
	}
	feedquery.ApplyViewerLikesToPage(view, posts)
}

func (s *Server) sessionClient(r *http.Request) *authoauth.SessionClient {
	if s.auth == nil {
		return nil
	}
	_, client, err := s.ResumeActiveClient(r)
	if err != nil {
		return nil
	}
	return client
}

func (s *Server) enrichProfileView(r *http.Request, view *profile.ProfileView) {
	if view == nil {
		return
	}
	s.enrichFeedLikes(r, &view.Feed)
}

func (s *Server) enrichTagView(r *http.Request, view *tag.TagView) {
	if view == nil {
		return
	}
	s.enrichFeedLikes(r, &view.Feed)
}
