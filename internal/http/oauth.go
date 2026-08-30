package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/simbachu/twisky/internal/actor"
	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	"github.com/simbachu/twisky/internal/auth/session"
	loginpage "github.com/simbachu/twisky/internal/components/login"
	"github.com/simbachu/twisky/internal/components/ui"
	profilequery "github.com/simbachu/twisky/internal/query/profile"
)

func (s *Server) accountMenuView(w http.ResponseWriter, r *http.Request, known ...ui.AuthorInfo) ui.AccountMenuView {
	if s.auth == nil {
		return ui.AccountMenuView{}
	}
	view := ui.AccountMenuView{Enabled: true}
	state, err := s.auth.Jar.Load(r)
	if err == nil {
		s.seedIdentityFromSession(state)
		state = s.syncSessionAuthorChrome(w, state, known...)
		return accountMenuViewFromState(state)
	}
	return view
}

func (s *Server) seedIdentityFromSession(state session.State) {
	if s.identity == nil {
		return
	}
	for _, account := range state.Accounts {
		if account.Handle != "" && account.Handle != "handle.invalid" && account.DID != "" {
			s.identity.Observe(account.Handle, account.DID)
		}
	}
}

// syncSessionAuthorChrome updates cookied accounts from AuthorInfo already loaded for the page.
// Saves only when chrome fields change. Returns the state to use for this response's menu.
func (s *Server) syncSessionAuthorChrome(w http.ResponseWriter, state session.State, known ...ui.AuthorInfo) session.State {
	if s.auth == nil {
		return state
	}
	changed := false
	for _, author := range known {
		if author.DID == "" {
			continue
		}
		next, ok := state.UpdateAuthorChrome(author.DID, author.Handle, author.DisplayName, author.Avatar)
		if !ok {
			continue
		}
		state = next
		changed = true
	}
	if !changed {
		return state
	}
	if err := s.auth.Jar.Save(w, state); err != nil {
		slog.Warn("session author chrome save failed", "err", err)
	}
	return state
}

func accountMenuViewFromState(state session.State) ui.AccountMenuView {
	view := ui.AccountMenuView{Enabled: true}
	active, ok := state.ActiveAccount()
	if !ok {
		return view
	}
	current := authorInfoFromSession(active)
	view.Current = &current
	for _, account := range state.Accounts {
		if account.DID == active.DID {
			continue
		}
		view.Additional = append(view.Additional, authorInfoFromSession(account))
	}
	return view
}

func authorInfoFromSession(account session.Account) ui.AuthorInfo {
	return ui.AuthorInfo{
		Handle:      account.Handle,
		DisplayName: actor.Name(account.DisplayName, account.Handle),
		DID:         account.DID,
		Avatar:      account.Avatar,
	}
}

func authorInfoFromProfileView(view profilequery.ProfileView) ui.AuthorInfo {
	return ui.AuthorInfo{
		Handle:      view.Handle,
		DisplayName: view.DisplayName,
		DID:         view.DID,
		Avatar:      view.Avatar,
		IsLabeler:   view.IsLabeler,
	}
}

func (s *Server) requireAuth() bool {
	return s.auth != nil
}

func (s *Server) handleClientMetadata(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}
	meta, err := s.auth.ClientMetadata()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = loginpage.Page("", s.publicBaseURL, "/oauth/login", s.hasCookiedAccounts(r)).Render(w)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	username := strings.TrimSpace(r.PostFormValue("username"))
	adding := s.hasCookiedAccounts(r)
	if username == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = loginpage.Page("Handle or DID is required.", s.publicBaseURL, "/oauth/login", adding).Render(w)
		return
	}

	startID := username
	if slug, err := actor.ParseSlug(username); err == nil && s.identity != nil {
		if did, err := s.identity.Resolve(r.Context(), slug, s.profileReader); err == nil && did != "" {
			startID = did
		}
	}

	redirectURL, err := s.auth.App.StartAuthFlow(r.Context(), startID)
	if err != nil {
		slog.Warn("oauth login failed", "err", err)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = loginpage.Page(fmt.Sprintf("Login failed: %v", err), s.publicBaseURL, "/oauth/login", adding).Render(w)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (s *Server) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}

	sessData, err := s.auth.App.ProcessCallback(r.Context(), r.URL.Query())
	if err != nil {
		http.Error(w, fmt.Sprintf("OAuth callback failed: %v", err), http.StatusBadRequest)
		return
	}

	handle := ""
	if s.identity != nil {
		if ident, err := s.identity.LookupDID(r.Context(), sessData.AccountDID); err == nil && !ident.Handle.IsInvalidHandle() {
			handle = ident.Handle.String()
		}
	}

	state, _ := s.auth.Jar.Load(r)
	if handle == "" || handle == "handle.invalid" {
		for _, account := range state.Accounts {
			if account.DID == sessData.AccountDID.String() && account.Handle != "" && account.Handle != "handle.invalid" {
				handle = account.Handle
				break
			}
		}
	}
	if s.identity != nil && handle != "" && handle != "handle.invalid" {
		s.identity.Observe(handle, sessData.AccountDID.String())
	}

	state = state.AddAccount(session.Account{
		DID:       sessData.AccountDID.String(),
		SessionID: sessData.SessionID,
		Handle:    handle,
	})
	if err := s.auth.Jar.Save(w, state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectTo := "/"
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

func (s *Server) handleOAuthLogout(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state, err := s.auth.Jar.Load(r)
	if err == nil {
		if account, ok := state.ActiveAccount(); ok {
			did, parseErr := syntax.ParseDID(account.DID)
			if parseErr == nil {
				if logoutErr := s.auth.App.Logout(r.Context(), did, account.SessionID); logoutErr != nil {
					slog.Warn("oauth logout revoke failed", "did", account.DID, "err", logoutErr)
				}
			}
			state = state.RemoveAccount(account.DID)
		}
		if len(state.Accounts) == 0 {
			s.auth.Jar.Clear(w)
		} else if saveErr := s.auth.Jar.Save(w, state); saveErr != nil {
			http.Error(w, saveErr.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		s.auth.Jar.Clear(w)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleOAuthSwitch(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth() {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	did := strings.TrimSpace(r.PostFormValue("did"))
	if did == "" {
		http.Error(w, "did is required", http.StatusBadRequest)
		return
	}

	state, err := s.auth.Jar.Load(r)
	if err != nil {
		http.Error(w, "not signed in", http.StatusBadRequest)
		return
	}
	next, ok := state.SetActive(did)
	if !ok {
		http.Error(w, "unknown account", http.StatusBadRequest)
		return
	}

	account, _ := next.ActiveAccount()
	parsed, parseErr := syntax.ParseDID(account.DID)
	if parseErr != nil {
		http.Error(w, "unknown account", http.StatusBadRequest)
		return
	}
	if _, err := s.auth.App.ResumeSession(r.Context(), parsed, account.SessionID); err != nil {
		slog.Warn("oauth switch resume failed", "did", account.DID, "err", err)
		s.invalidateOAuthSession(&account)
		state = state.RemoveAccount(account.DID)
		if len(state.Accounts) == 0 {
			s.auth.Jar.Clear(w)
		} else if saveErr := s.auth.Jar.Save(w, state); saveErr != nil {
			http.Error(w, saveErr.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, sameOriginRedirect(r), http.StatusFound)
		return
	}

	if err := s.auth.Jar.Save(w, next); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, sameOriginRedirect(r), http.StatusFound)
}

func (s *Server) hasCookiedAccounts(r *http.Request) bool {
	if s.auth == nil {
		return false
	}
	state, err := s.auth.Jar.Load(r)
	return err == nil && len(state.Accounts) > 0
}

func sameOriginRedirect(r *http.Request) string {
	ref := strings.TrimSpace(r.Referer())
	if ref == "" {
		return "/"
	}
	parsed, err := url.Parse(ref)
	if err != nil || parsed.Path == "" || strings.HasPrefix(parsed.Path, "//") {
		return "/"
	}
	if parsed.IsAbs() && parsed.Host != r.Host {
		return "/"
	}
	if parsed.RawQuery != "" {
		return parsed.Path + "?" + parsed.RawQuery
	}
	return parsed.Path
}

// ResumeActiveSession returns the active browser account after validating the OAuth session.
func (s *Server) ResumeActiveSession(r *http.Request) (*session.Account, error) {
	account, err := s.loadActiveAccount(r)
	if err != nil {
		return nil, err
	}
	did, err := syntax.ParseDID(account.DID)
	if err != nil {
		return nil, err
	}
	if _, err := s.auth.App.ResumeSession(r.Context(), did, account.SessionID); err != nil {
		return nil, err
	}
	return account, nil
}

// ResumeActiveClient resumes the active OAuth session and returns an authenticated API client.
func (s *Server) ResumeActiveClient(r *http.Request) (*session.Account, *authoauth.SessionClient, error) {
	return s.cachedResumeActiveClient(r)
}

func (s *Server) resumeActiveClientUncached(r *http.Request, account *session.Account) (*session.Account, *authoauth.SessionClient, error) {
	did, err := syntax.ParseDID(account.DID)
	if err != nil {
		return nil, nil, err
	}
	sess, err := s.auth.App.ResumeSession(r.Context(), did, account.SessionID)
	if err != nil {
		return nil, nil, err
	}
	if sess == nil || sess.Data == nil {
		return nil, nil, fmt.Errorf("oauth: incomplete session")
	}
	return account, authoauth.NewSessionClient(sess.APIClient()), nil
}

func (s *Server) loadActiveAccount(r *http.Request) (*session.Account, error) {
	if s.auth == nil {
		return nil, session.ErrMissingCookie
	}
	state, err := s.auth.Jar.Load(r)
	if err != nil {
		return nil, err
	}
	account, ok := state.ActiveAccount()
	if !ok {
		return nil, session.ErrMissingCookie
	}
	return &account, nil
}

// clearStaleOAuthAccount drops the active browser account when the OAuth session
// can no longer be used (for example after invalid_grant on token refresh).
func (s *Server) clearStaleOAuthAccount(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		return
	}
	state, err := s.auth.Jar.Load(r)
	if err != nil {
		return
	}
	account, ok := state.ActiveAccount()
	if !ok {
		return
	}
	s.invalidateOAuthSession(&account)
	state = state.RemoveAccount(account.DID)
	if len(state.Accounts) == 0 {
		s.auth.Jar.Clear(w)
		return
	}
	_ = s.auth.Jar.Save(w, state)
}
