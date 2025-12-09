package main

import (
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/lib/loger/handlers/slogpretty"
	"syscall"
)

const (
	envLocal      = "local"
	envProduction = "prod"
	envDev        = "dev"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting server", slog.String("env", cfg.Env))

	application := app.New(log, cfg.Grpc.Port, cfg.StoragePath, cfg.TokenTTL)
	go application.GRPCServer.MustStart()
	// инициализировать приложение (app)
	// запустить приложение grpc сервер
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sigmessage := <-stop
	log.Info("stopping by signal", slog.String("signal", sigmessage.String()))
	application.GRPCServer.Stop()
	log.Info("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case "local":
		log = setupPrettySlog()
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "dev":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
	}
	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}
