package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) handleGetPosts(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("userId")
	if user == "" {
		http.Error(w, "userId required", http.StatusBadRequest)
		return
	}
	uid, err := strconv.Atoi(user)
	if err != nil {
		http.Error(w, "invalid userId", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	items, err := s.api.QueryByUser(ctx, uid)
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items": items,
	})
}
