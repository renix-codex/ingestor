package http

import (
	"encoding/json"
	"net/http"
)

// Register the GET /posts route
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		status := s.api.Health()
		resp := map[string]any{"status": status}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	s.mux.HandleFunc("GET /posts", s.handleGetPosts)
}
