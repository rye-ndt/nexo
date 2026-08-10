package main

import (
	"embed"
	"log"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app, err := wire(assets)
	if err != nil {
		log.Fatalf("wire: %v", err)
	}

	if err := app.AppBuilder.Run(); err != nil {
		app.Logger.Error("app run", "err", err)
	}
}
