package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/sirupsen/logrus"
	"github.com/t-mutaguchi-10antz/go-encoding" // プライベートリポジトリ
)

var (
	log *logrus.Logger
	srv *http.Server
)

const (
	gracefulDuration = 10 * time.Second
)

func init() {
	initLog()
	initSrv()
}

func main() {
	g := run.Group{}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	cancel := make(chan struct{})

	g.Add(
		func() error {
			select {
			case <-sig:
			case <-cancel:
			}
			return nil
		},
		func(err error) {
			close(cancel)
		},
	)

	g.Add(
		func() error {
			log.Infof("start serving on %s", srv.Addr)

			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Errorf("failed to serve: %v", err)
				return fmt.Errorf("failed to serve: %w", err)
			}

			return nil
		},
		func(err error) {
			log.Infof("gracefully shutting down...")

			ctx, cancel := context.WithTimeout(context.Background(), gracefulDuration)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				log.Fatalf("forced shutdown: %+v", err)
			}

			log.Infof("graceful shutdown")
		},
	)

	if err := g.Run(); err != nil {
		os.Exit(1)
	}
}

func initSrv() {
	var port string
	if v := os.Getenv("PORT"); v != "" {
		port = v
	} else {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		dbAvailable := true
		if dbAvailable {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)

		reqID := encoding.UUID()

		log.Infof("x-request-id: %s", reqID)
		w.Header().Set("x-request-id", reqID)

		_, _ = w.Write([]byte(fmt.Sprintf("Hello! RequestID: %s\n", reqID)))
	})

	srv = &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", port),
		Handler: mux,
	}
}

func initLog() {
	log = &logrus.Logger{}
	fieldMap := logrus.FieldMap{
		logrus.FieldKeyTime:  "timestamp",
		logrus.FieldKeyLevel: "severity",
		logrus.FieldKeyMsg:   "message",
	}

	switch os.Getenv("LOG_FORMAT") {
	case "json":
		log.Formatter = &logrus.JSONFormatter{FieldMap: fieldMap, TimestampFormat: time.RFC3339Nano}
	default:
		log.Formatter = &logrus.TextFormatter{FieldMap: fieldMap, TimestampFormat: time.RFC3339Nano}
	}

	switch os.Getenv("LOG_LEVEL") {
	case "trace":
		log.Level = logrus.TraceLevel
		log.Out = os.Stdout
	case "info":
		log.Level = logrus.InfoLevel
		log.Out = os.Stdout
	case "debug":
		log.Level = logrus.DebugLevel
		log.Out = os.Stdout
	case "warn", "warning":
		log.Level = logrus.WarnLevel
		log.Out = os.Stderr
	case "error":
		log.Level = logrus.ErrorLevel
		log.Out = os.Stderr
	case "fatal":
		log.Level = logrus.FatalLevel
		log.Out = os.Stderr
	default:
		log.Level = logrus.PanicLevel
		log.Out = os.Stderr
	}
}
