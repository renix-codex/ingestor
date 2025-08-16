package http

import (
	"context"
	"net/http"

	"github.com/renix-codex/ingestor/internal/api"
)

type Server struct {
	api *api.API
	mux *http.ServeMux
}

func New(a *api.API) *Server {
	s := &Server{api: a, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}
	go func() {
		<-ctx.Done()
		_ = httpSrv.Shutdown(context.Background())
	}()
	return httpSrv.ListenAndServe()
}
