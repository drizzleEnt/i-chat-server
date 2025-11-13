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

func (a *App) Run() error {
	// Start the application logic here

	go func() {
		if err := a.StartHttpServer(); err != nil && err != http.ErrServerClosed {
			a.serviceProvider.Logger(context.Background()).Error("error start http server", zap.Error(err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	a.serviceProvider.Logger(context.Background()).Info("shutting down the server...")
	if err := a.chatServer.Shutdown(context.Background()); err != nil {
		fmt.Printf("(err == http.ErrServerClosed): %v\n", (err == http.ErrServerClosed))
		a.serviceProvider.Logger(context.Background()).Error("error shutting down the server", zap.Error(err))
	}
	a.serviceProvider.Logger(context.Background()).Info("server shut down successfully")

	return nil
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

	a.serviceProvider = sp
	return nil
}

func (a *App) initHttpServer(ctx context.Context) error {
	muxRouter := routes.InitRoutes(a.serviceProvider.ChatController(ctx))

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

func (a *App) StartHttpServer() error {
	a.serviceProvider.Logger(context.Background()).Info("HTTP server is running", zap.String("address", a.serviceProvider.HttpConfig().Address()))
	if err := a.chatServer.ListenAndServe(); err != nil {
		return err
	}

	return nil
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
