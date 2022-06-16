package main

import (
	"context"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	addr := "localhost:8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	logger := setupLogging()
	defer logger.Sync()

	server := newServer(logger, addr)

	runningCtx, runningCancel := context.WithCancel(ctx)
	defer runningCancel()
	go func() {
		logger.Info("starting server", zap.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("starting failed", zap.Error(err))
			runningCancel() // initiate shutdown sequence
		}
	}()

	<-runningCtx.Done()
	logger.Info("received termination signal, shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Minute)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Warn("failed to shutdown server", zap.Error(err))
	}
	logger.Info("server shutdown complete")
}

type Server struct {
	logger *zap.Logger
}

func newServer(logger *zap.Logger, addr string) *http.Server {
	srv := Server{logger: logger}
	return &http.Server{
		Addr:    addr,
		Handler: srv.handler(),
	}
}

func (s *Server) handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/slow/{duration}", s.slow)
	return r
}

func (s *Server) slow(rw http.ResponseWriter, req *http.Request) {
	logger := s.logger.With(
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
	)

	duration := mux.Vars(req)["duration"]
	if duration == "" {
		logger.Info("using default duration")
		duration = "10s"
	}

	pause, err := time.ParseDuration(duration)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to parse duration")
		rw.WriteHeader(http.StatusBadRequest)
	}

	logger.Info("starting request")
	logger.Sugar().Infof("pausing for %s", pause)
	time.Sleep(pause)
	logger.Info("finishing request")

	return
}

func setupLogging() *zap.Logger {
	conf := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:       false,
		Encoding:          "json",
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		DisableStacktrace: true,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}
	logger, err := conf.Build()
	if err != nil {
		panic(err)
	}
	return logger
}
