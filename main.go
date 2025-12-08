package main

import (
	"log"

	"github.com/ogpourya/dnsbro/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("dnsbro: %v", err)
	}
}
