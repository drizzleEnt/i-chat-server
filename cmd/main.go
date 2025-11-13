package main

import (
	"chatsrv/internal/app"
	"context"
	"log"
)

func main() {
	ctx := context.Background()
	app, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to initialize app: %s", err.Error())
	}

	if err := app.Run(); err != nil {
		log.Fatalf("application error: %s", err.Error())
	}
}
