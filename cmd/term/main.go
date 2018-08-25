package main

import (
	"context"
	"log"

	"github.com/qsydev/goterm/internal/terminal"
)

func main() {
	// TODO: check for setup stuff
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t := &terminal.T{}
	if err := t.Run(ctx); err != nil {
		log.Printf("terminal interrupted: %s", err)
	}
}
