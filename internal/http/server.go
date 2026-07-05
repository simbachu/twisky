package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	postpage "github.com/simbachu/twisky/internal/components/post"
	profilepage "github.com/simbachu/twisky/internal/components/profile"
	tagpage "github.com/simbachu/twisky/internal/components/tag"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/tag"
	"github.com/simbachu/twisky/internal/response"
)

type Server struct {
	queries *query.Dispatcher
}

func NewServer(queries *query.Dispatcher) *Server {
	return &Server{queries: queries}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{slug}/post/{id}", s.handlePost)
	mux.HandleFunc("GET /{head}/{tail}", s.handleTwoSegment)
	mux.HandleFunc("GET /{slug}", s.handleProfile(intent.ProfileTabPosts))
	return mux
}

func (s *Server) handleTwoSegment(w http.ResponseWriter, r *http.Request) {
	head := r.PathValue("head")
	tail := r.PathValue("tail")
	if head == "tagged" {
		s.dispatchTag(w, r, tail)
		return
	}
	if tail == "media" {
		r.SetPathValue("slug", head)
		s.handleProfile(intent.ProfileTabMedia)(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) dispatchTag(w http.ResponseWriter, r *http.Request, tagName string) {
	resp, err := s.queries.Dispatch(r.Context(), intent.ViewTag{Tag: tagName})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	switch v := resp.(type) {
	case tag.TagView:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tagpage.Tag(v).Render(w)
	case response.ErrorResponse:
		http.Error(w, v.Message, v.Status)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (s *Server) handleProfile(tab intent.ProfileTab) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		resp, err := s.queries.Dispatch(r.Context(), intent.ViewProfile{Slug: slug, Tab: tab})
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		switch v := resp.(type) {
		case profile.ProfileView:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = profilepage.Profile(v).Render(w)
		case response.ErrorResponse:
			http.Error(w, v.Message, v.Status)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	postID, err := url.PathUnescape(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	resp, err := s.queries.Dispatch(r.Context(), intent.ViewPost{
		Slug: r.PathValue("slug"),
		ID:   postID,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	switch v := resp.(type) {
	case feedquery.PostPageView:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = postpage.PostPage(v).Render(w)
	case response.ErrorResponse:
		http.Error(w, v.Message, v.Status)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func ListenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return server.Shutdown(context.Background())
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("http server: %w", err)
	}
}
