package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avitotech/tenders/internal/config"
	"avitotech/tenders/internal/lib/logger/sl"
	"avitotech/tenders/internal/lib/logger/slogpretty"
	"avitotech/tenders/internal/storage/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envlocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// TODO : init config : cleanenv
	Cfg := config.MustLoad()

	fmt.Println(Cfg)

	// TODO : init logger : slog
	log := setupLogger("local")
	log.Info("starting service", slog.String("env", "local"))
	log.Debug("debug message are enabled")

	// TODO : init db : Postresql(sqlite)
	storage, err := postgres.New(Cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	// TODO : init router : chi

	r := chi.NewRouter()
	// middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	/*r.Use(middleware.BasicAuth("/", map[string]string{
		"admin": "admin",
	})) */
	r.Get("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		storage.Ping()
		_, err := w.Write([]byte("ok"))
		if err != nil {
			log.Info("error response")
		}
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ti loh ahahhhahhahhahhaah"))
		if err != nil {
			log.Info("error response")
		}
	})

	log.Info("starting server", slog.String("address", Cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	srv := &http.Server{
		Addr:         Cfg.Address,
		Handler:      r,
		ReadTimeout:  Cfg.HttpServer.Timeout,
		WriteTimeout: Cfg.HttpServer.Timeout,
		IdleTimeout:  Cfg.HttpServer.Idle_timeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: close storage

	log.Info("server stopped")

	// TODO : run server
	/*r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Get("/loh", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("nu ti loh konesno"))
		fmt.Println("Serving", " ", r.URL, " ", r.Host)

	})

	http.ListenAndServe(":3000", r)
	*/

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envlocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
