package http

import (
	"net/http"
	"strings"

	"github.com/simbachu/twisky/internal/actor"
	"github.com/simbachu/twisky/internal/atproto"
	composepage "github.com/simbachu/twisky/internal/components/compose"
	loginpage "github.com/simbachu/twisky/internal/components/login"
	"github.com/simbachu/twisky/internal/intent"
)

func (s *Server) handleComposeNew(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}
	if _, err := s.loadActiveAccount(r); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = loginpage.Page("", s.publicBaseURL, "/my/posts/new", false).Render(w)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	suggested := s.suggestedAccounts(r.Context())
	_ = composepage.NewPostPage("", "", s.publicBaseURL, suggested, s.accountMenuView(w, r, suggested...)).Render(w)
}

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	text := r.FormValue("text")

	account, err := s.loadActiveAccount(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	writer := s.postWriter
	if writer == nil {
		_, client, err := s.ResumeActiveClient(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		writer = client
	}
	if s.commands == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	recordURI, err := s.commands.Dispatch(r.Context(), intent.CreatePost{Text: text}, writer)
	if err != nil {
		message := composeErrorMessage(err)
		if message != "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			suggested := s.suggestedAccounts(r.Context())
			_ = composepage.NewPostPage(message, text, s.publicBaseURL, suggested, s.accountMenuView(w, r, suggested...)).Render(w)
			return
		}
		http.Error(w, "post failed", http.StatusBadGateway)
		return
	}

	parsed, err := atproto.ParsePostURI(recordURI)
	if err != nil {
		http.Error(w, "post failed", http.StatusBadGateway)
		return
	}
	location := actor.PostPath(account.Handle, account.DID, parsed.Rkey())
	http.Redirect(w, r, location, http.StatusSeeOther)
}

func composeErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	switch {
	case strings.Contains(msg, "text is required"):
		return "Post text is required."
	case strings.Contains(msg, "exceeds"):
		return "Post text is too long."
	default:
		return ""
	}
}
