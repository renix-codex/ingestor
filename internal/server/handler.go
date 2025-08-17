package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) handleGetPosts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	user := q.Get("userId")
	limit := parseInt(q.Get("limit"), 50)
	offset := parseInt(q.Get("offset"), 0)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var (
		items any
		err   error
	)

	if user == "" {
		// no userId provided -> recent with pagination
		items, err = s.api.QueryRecent(ctx, limit, offset)
	} else {
		uid, convErr := strconv.Atoi(user)
		if convErr != nil {
			http.Error(w, "invalid userId", http.StatusBadRequest)
			return
		}
		items, err = s.api.QueryByUser(ctx, uid)
	}
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

// parseInt parses a string to int, returning def if parsing fails.
// parseInt returns def when s is empty or not an int.
func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
