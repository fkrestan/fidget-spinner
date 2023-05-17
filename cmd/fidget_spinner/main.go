package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/fkrestan/fidget_spinner/internal/handler"
	"github.com/fkrestan/fidget_spinner/internal/middleware"
)

func main() {
	svcLogger, svcAtomicLevel, err := setupServiceLogger()
	defer svcLogger.Sync()
	l := svcLogger.Sugar()

	managementPort := getEnv("MGMT_SERVER_PORT", "9090")
	managementAddr := fmt.Sprintf(":%s", managementPort)
	publicPort := getEnv("TIME_SERVER_PORT", "8080")
	publicAddr := fmt.Sprintf(":%s", publicPort)

	// Access logging
	accessLogger, accessAtomicLevel, err := setupAccessLogger()
	if err != nil {
		l.Fatalw("Error configuring Zap access logger", "error", err)
	}
	defer accessLogger.Sync()

	promRegisterer := prometheus.DefaultRegisterer
	promGatherer := prometheus.DefaultGatherer
	promHandler := promhttp.HandlerFor(promGatherer, promhttp.HandlerOpts{
		Registry: promRegisterer,
		ErrorLog: zap.NewStdLog(svcLogger),
	})

	// Management API
	managementMux := http.NewServeMux()
	managementMux.Handle("/metrics", promHandler)
	managementMux.HandleFunc("/livez", handler.Liveness)
	managementMux.HandleFunc("/servicelog", svcAtomicLevel.ServeHTTP)
	managementMux.HandleFunc("/accesslog", accessAtomicLevel.ServeHTTP)

	managementServer := &http.Server{
		Addr:         managementAddr,
		ErrorLog:     zap.NewStdLog(svcLogger),
		Handler:      managementMux,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		l.Infow("Starting management API HTTP server", "addr", managementAddr)
		err = managementServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			l.Errorw("Management server error", "error", err)
		}

		l.Infow("Management server stopped")
	}()

	// Public API
	publicAPIRouter := http.NewServeMux()
	publicAPIRouter.Handle("/spin", &handler.SpinHandler{L: l})

	mLogging := middleware.Logging(accessLogger)
	mMetrics := middleware.HTTPMetrics(promRegisterer)

	publicAPIServer := &http.Server{
		Addr:         publicAddr,
		ErrorLog:     zap.NewStdLog(svcLogger),
		Handler:      mLogging(mMetrics(publicAPIRouter)),
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		IdleTimeout:  5 * time.Second,
	}

	go func() {
		l.Infow("Starting public API HTTP server", "addr", publicAddr)
		err = publicAPIServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			l.Errorw("Public API server error", "error", err)
		}

		l.Infow("Public API server stopped")
	}()

	// Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	l.Info("Stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := publicAPIServer.Shutdown(ctx); err != nil {
			l.Fatalw("Error stopping public API server", "error", err)
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		if err := managementServer.Shutdown(ctx); err != nil {
			l.Fatalw("Error stopping management server", "error", err)
		}
		wg.Done()
	}()
	wg.Wait()
}

func setupServiceLogger() (*zap.Logger, *zap.AtomicLevel, error) {
	atomicLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	loggerConfig := zap.Config{
		Level:             atomicLevel,
		DisableCaller:     false,
		DisableStacktrace: false,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.NanosDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("build zap service logger: %w", err)
	}

	return logger, &atomicLevel, nil
}

func setupAccessLogger() (*zap.Logger, *zap.AtomicLevel, error) {
	atomicLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	// TODO sampler
	loggerConfig := zap.Config{
		Level:             atomicLevel,
		DisableCaller:     true,
		DisableStacktrace: true,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        zapcore.OmitKey,
			CallerKey:      zapcore.OmitKey,
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     zapcore.OmitKey,
			StacktraceKey:  zapcore.OmitKey,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.NanosDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("build zap access logger: %w", err)
	}

	return logger, &atomicLevel, nil
}

func getEnv(name, defaultVaule string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		v = defaultVaule
	}
	return v
}
