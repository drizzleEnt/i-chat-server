package app

import (
	"chatsrv/internal/routes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type App struct {
	serviceProvider *serviceProvider
	chatServer      *http.Server
}

func New(ctx context.Context) (*App, error) {
	a := &App{}

	if err := a.initDependencies(ctx); err != nil {
		log.Printf("error initialize dependencies %s", err.Error())
		return nil, err
	}
	// Initialize app components here
	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	nctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() { serverErr <- a.StartHttpServer(nctx) }()

	var runErr error
	select {
	case <-nctx.Done():
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			runErr = err
		}
	}

	a.serviceProvider.Logger(ctx).Info("shutting down the server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := a.chatServer.Shutdown(shutdownCtx); err != nil {
		a.serviceProvider.Logger(ctx).Error("error shutting down the server", zap.Error(err))
		if runErr == nil {
			runErr = err
		}
	}
	if err := a.serviceProvider.DBClient().Close(); err != nil && runErr == nil {
		runErr = err
	}
	if a.serviceProvider.pool != nil {
		a.serviceProvider.pool.Close()
	}
	a.serviceProvider.Logger(ctx).Info("server shut down successfully")

	return runErr
}

func (a *App) initDependencies(ctx context.Context) error {
	// Initialize dependencies here
	deps := []func(context.Context) error{
		a.initServiceProvider,
		a.initHttpServer,
	}

	for _, v := range deps {
		if err := v(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initServiceProvider(ctx context.Context) error {
	sp := newServiceProvider()
	sp.Logger(ctx)
	if err := sp.initDB(ctx); err != nil {
		return fmt.Errorf("initialize postgres connection: %w", err)
	}

	a.serviceProvider = sp
	return nil
}

func (a *App) initHttpServer(ctx context.Context) error {
	muxRouter := routes.InitRoutes(a.serviceProvider.ChatController(ctx), a.serviceProvider.DBClient().PingContext)

	a.chatServer = &http.Server{
		Addr:           a.serviceProvider.HttpConfig().Address(),
		Handler:        muxRouter,
		ReadTimeout:    10 * 1e9,
		WriteTimeout:   10 * 1e9,
		IdleTimeout:    120 * 1e9,
		MaxHeaderBytes: 1 << 20,
	}

	// Initialize HTTP server here
	return nil
}

func (a *App) StartHttpServer(ctx context.Context) error {
	a.serviceProvider.Logger(ctx).Info("HTTP server is running", zap.String("address", a.serviceProvider.HttpConfig().Address()))
	return a.chatServer.ListenAndServe()
}

func getCore(level zap.AtomicLevel) zapcore.Core {
	stdout := zapcore.AddSync(os.Stdout)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	develomentCfg := zap.NewDevelopmentEncoderConfig()
	develomentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(develomentCfg)

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
	)
}

var logLevel = flag.String("l", "info", "log level")

func getAtomicLevel() zap.AtomicLevel {
	var level zapcore.Level

	if err := level.Set(*logLevel); err != nil {
		log.Fatalf("failed to set log level: %v", err)
	}

	return zap.NewAtomicLevelAt(level)
}
