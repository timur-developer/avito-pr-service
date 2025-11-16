package server

import (
	"avito-pr-service/config"
	"avito-pr-service/internal/handler"
	"avito-pr-service/internal/repository/postgres"
	"avito-pr-service/internal/usecase"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
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

	teamRepository := store.Team()
	userRepository := store.User()
	prRepository := store.PR()

	userUC := usecase.NewUserUsecase(userRepository, log)
	prUC := usecase.NewPRUsecase(prRepository, userRepository, teamRepository, log)
	teamUC := usecase.NewTeamUsecase(teamRepository, userRepository, prRepository, prUC, log)

	teamHandler := handler.NewTeamHandler(teamUC, log)
	userHandler := handler.NewUserHandler(userUC, log)
	prHandler := handler.NewPRHandler(prUC, log)

	r := chi.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(c.Handler)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// редирект на сваггер
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://localhost:8081", http.StatusFound)
	})

	teamHandler.Register(r)
	userHandler.Register(r)
	prHandler.Register(r)

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
