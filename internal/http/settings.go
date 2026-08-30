package http

import (
	"io"
	"net/http"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
	"github.com/simbachu/twisky/internal/bluesky"
	settingscommand "github.com/simbachu/twisky/internal/command/settings"
	loginpage "github.com/simbachu/twisky/internal/components/login"
	settingspage "github.com/simbachu/twisky/internal/components/settings"
	"github.com/simbachu/twisky/internal/intent"
	settingsquery "github.com/simbachu/twisky/internal/query/settings"
	"github.com/simbachu/twisky/internal/response"
)

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}

	account, reader, ok := s.settingsPageReader(r)
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = loginpage.Page("", s.publicBaseURL, "/settings", false).Render(w)
		return
	}

	resp := settingsquery.NewHandler(reader).Handle(r.Context(), intent.ViewSettings{
		DID:    account.DID,
		Handle: account.Handle,
	})
	switch v := resp.(type) {
	case settingsquery.SettingsView:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		suggested := s.suggestedAccounts(r.Context())
		_ = settingspage.Settings(v, suggested, s.accountMenuView(w, r, suggested...), s.publicBaseURL).Render(w)
	case response.ErrorResponse:
		if v.Status == http.StatusUnauthorized {
			s.clearStaleOAuthAccount(w, r)
		}
		s.writeQueryError(w, r, v)
	default:
		s.writeQueryError(w, r, response.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Something went wrong loading this page",
		})
	}
}

func (s *Server) settingsPageReader(r *http.Request) (*session.Account, settingsquery.Reader, bool) {
	if s.settingsReader != nil {
		account, err := s.loadActiveAccount(r)
		if err != nil {
			return nil, nil, false
		}
		return account, s.settingsReader, true
	}
	account, client, err := s.ResumeActiveClient(r)
	if err != nil {
		return nil, nil, false
	}
	return account, client, true
}

func (s *Server) settingsPageWriter(r *http.Request) (settingscommand.Writer, error) {
	if s.settingsWriter != nil {
		if _, err := s.loadActiveAccount(r); err != nil {
			return nil, err
		}
		return s.settingsWriter, nil
	}
	_, client, err := s.ResumeActiveClient(r)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (s *Server) handleUpdateContentFiltering(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	writer, err := s.settingsPageWriter(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	labels := make(map[string]string, len(bluesky.ContentFilterLabels))
	for _, identifier := range bluesky.ContentFilterLabels {
		labels[identifier] = r.FormValue(identifier)
	}
	prefs, err := settingscommand.NewHandler(writer).HandleContentFiltering(r.Context(), intent.UpdateContentFiltering{
		AdultContentEnabled: r.FormValue("adult_content") == "1",
		Labels:              labels,
	})
	if err != nil {
		s.writeSettingsWriteError(w, r, err)
		return
	}
	s.writeSettingsSection(w, r, settingspage.ContentFilteringFragment(prefs))
}

func (s *Server) handleUpdateThreading(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	writer, err := s.settingsPageWriter(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	prefs, err := settingscommand.NewHandler(writer).HandleThreading(r.Context(), intent.UpdateThreading{
		Sort:                    r.FormValue("sort"),
		PrioritizeFollowedUsers: r.FormValue("prioritize_followed") == "1",
	})
	if err != nil {
		s.writeSettingsWriteError(w, r, err)
		return
	}
	s.writeSettingsSection(w, r, settingspage.ThreadingFragment(prefs))
}

func (s *Server) writeSettingsWriteError(w http.ResponseWriter, r *http.Request, err error) {
	if authoauth.IsInsufficientAuth(err) {
		s.writeQueryError(w, r, response.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "Your Bluesky login needs to be refreshed. Log out and sign in again to save settings.",
		})
		return
	}
	s.writeQueryError(w, r, response.ErrorResponse{
		Status:  http.StatusBadGateway,
		Message: "Failed to save settings",
	})
}

func (s *Server) writeSettingsSection(w http.ResponseWriter, r *http.Request, section interface{ Render(io.Writer) error }) {
	if !isHTMXRequest(r) {
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = section.Render(w)
}
