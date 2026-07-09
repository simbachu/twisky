package http

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	postpage "github.com/simbachu/twisky/internal/components/post"
	profilepage "github.com/simbachu/twisky/internal/components/profile"
	tagpage "github.com/simbachu/twisky/internal/components/tag"
	"github.com/simbachu/twisky/internal/intent"
	"github.com/simbachu/twisky/internal/query"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
	"github.com/simbachu/twisky/internal/query/profile"
	"github.com/simbachu/twisky/internal/query/tag"
	"github.com/simbachu/twisky/internal/response"
	"github.com/simbachu/twisky/static"
)

type Server struct {
	queries *query.Dispatcher
}

func NewServer(queries *query.Dispatcher) *Server {
	return &Server{queries: queries}
}

func (s *Server) Handler() http.Handler {
	staticFS, err := fs.Sub(static.WebFS, "web")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServerFS(staticFS)))
	r.Get("/tagged/{tag}", s.handleTag)
	r.Get("/{slug}/post/{id}", s.handlePost)
	r.Get("/{slug}/media", s.handleProfile(intent.ProfileTabMedia))
	r.Get("/{slug}", s.handleProfile(intent.ProfileTabPosts))
	return r
}

func (s *Server) handleTag(w http.ResponseWriter, r *http.Request) {
	s.dispatchTag(w, r, chi.URLParam(r, "tag"))
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
		_ = tagpage.Tag(v, time.Now().UTC()).Render(w)
	case response.ErrorResponse:
		http.Error(w, v.Message, v.Status)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (s *Server) handleProfile(tab intent.ProfileTab) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		resp, err := s.queries.Dispatch(r.Context(), intent.ViewProfile{Slug: slug, Tab: tab})
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		switch v := resp.(type) {
		case profile.ProfileView:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = profilepage.Profile(v, time.Now().UTC()).Render(w)
		case response.ErrorResponse:
			http.Error(w, v.Message, v.Status)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	postID, err := url.PathUnescape(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	resp, err := s.queries.Dispatch(r.Context(), intent.ViewPost{
		Slug: chi.URLParam(r, "slug"),
		ID:   postID,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	switch v := resp.(type) {
	case feedquery.PostPageView:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = postpage.PostPage(v, time.Now().UTC()).Render(w)
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
