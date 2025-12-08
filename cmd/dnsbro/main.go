package main

import (
	"log"

	"dnsbro/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("dnsbro: %v", err)
	}
}
