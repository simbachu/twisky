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
	count := 0
	if raw := strings.TrimSpace(r.FormValue("count")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			count = n
		}
	}

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

	err := s.commands.Dispatch(r.Context(), intent.LikePost{URI: uri, CID: cid}, writer)
	if err != nil && !authoauth.IsDuplicateLike(err) {
		http.Error(w, "like failed", http.StatusBadGateway)
		return
	}

	view := feedquery.PostViewWithStrongRef(feedquery.PostView{
		Liked:     true,
		LikeCount: count + 1,
	}, uri, cid)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = postpage.LikeButtonFragment(view, "").Render(w)
}
