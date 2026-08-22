package http

import (
	"net/http"
	"strconv"
	"strings"

	authoauth "github.com/simbachu/twisky/internal/auth/oauth"
	postpage "github.com/simbachu/twisky/internal/components/post"
	"github.com/simbachu/twisky/internal/intent"
	feedquery "github.com/simbachu/twisky/internal/query/feed"
)

func (s *Server) handleLike(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	uri := strings.TrimSpace(r.FormValue("uri"))
	cid := strings.TrimSpace(r.FormValue("cid"))
	if uri == "" || cid == "" {
		http.Error(w, "uri and cid are required", http.StatusBadRequest)
		return
	}
	count := parseEngagementCount(r.FormValue("count"))

	writer := s.likeWriter
	if writer == nil {
		_, client, err := s.ResumeActiveClient(r)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writer = client
	}
	if s.commands == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	recordURI, err := s.commands.Dispatch(r.Context(), intent.LikePost{URI: uri, CID: cid}, writer)
	if err != nil && !authoauth.IsDuplicateLike(err) {
		http.Error(w, "like failed", http.StatusBadGateway)
		return
	}

	view := feedquery.PostViewWithStrongRef(feedquery.PostView{
		ID:        strings.TrimSpace(r.FormValue("id")),
		Liked:     true,
		LikeCount: count + 1,
	}, uri, cid)
	if recordURI != "" {
		view = feedquery.PostViewWithEngagement(view, recordURI, "")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = postpage.LikeActionFragment(view).Render(w)
}

func (s *Server) handleUnlike(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	record := strings.TrimSpace(r.FormValue("record"))
	uri := strings.TrimSpace(r.FormValue("uri"))
	cid := strings.TrimSpace(r.FormValue("cid"))
	if record == "" {
		http.Error(w, "record is required", http.StatusBadRequest)
		return
	}
	count := parseEngagementCount(r.FormValue("count"))

	writer := s.likeWriter
	if writer == nil {
		_, client, err := s.ResumeActiveClient(r)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writer = client
	}
	if s.commands == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err := s.commands.Dispatch(r.Context(), intent.UnlikePost{Record: record}, writer)
	if err != nil && !authoauth.IsRecordNotFound(err) {
		http.Error(w, "unlike failed", http.StatusBadGateway)
		return
	}

	view := feedquery.PostViewWithStrongRef(feedquery.PostView{
		ID:        strings.TrimSpace(r.FormValue("id")),
		Liked:     false,
		LikeCount: decrementEngagementCount(count),
	}, uri, cid)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = postpage.LikeActionFragment(view).Render(w)
}

func parseEngagementCount(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func decrementEngagementCount(count int) int {
	if count <= 0 {
		return 0
	}
	return count - 1
}
