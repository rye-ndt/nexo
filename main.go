package main

import (
	"embed"
	"log"
	"os"
	"runtime/debug"
)

//go:embed all:frontend/dist
var assets embed.FS

const defaultMemoryLimit = 1 << 30

func main() {
	if os.Getenv("GOMEMLIMIT") == "" {
		debug.SetMemoryLimit(defaultMemoryLimit)
	}

	app, err := wire(assets)
	if err != nil {
		log.Fatalf("wire: %v", err)
	}

	if err := app.AppBuilder.Run(); err != nil {
		app.Logger.Error("app run", "err", err)
	}
}
