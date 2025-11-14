package server

import (
	"avito-pr-service/config"
	"avito-pr-service/internal/handler"
	"avito-pr-service/internal/repository/postgres"
	"avito-pr-service/internal/usecase"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

type Server struct {
	http  *http.Server
	store *postgres.Store
	log   *slog.Logger
}

func New(cfg config.Config) (*Server, error) {
	ctx := context.Background()

	store, err := postgres.NewStore(ctx, cfg.DSN)
	if err != nil {
		return nil, err
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With("service", "avito-pr-service")

	teamUC := usecase.NewTeamUsecase(store.Team(), log)
	userUC := usecase.NewUserUsecase(store.User(), log)

	teamHandler := handler.NewTeamHandler(teamUC, log)
	userHandler := handler.NewUserHandler(userUC, log)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	teamHandler.Register(r)
	userHandler.Register(r)

	httpSrv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	return &Server{
		http:  httpSrv,
		store: store,
		log:   log,
	}, nil
}

func (s *Server) Start() error {
	s.log.Info("server starting", "addr", s.http.Addr)
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down server...")

	if err := s.http.Shutdown(ctx); err != nil {
		s.log.Error("http shutdown error", "err", err)
	}

	s.store.Close()

	return nil
}
