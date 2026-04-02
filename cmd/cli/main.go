package main

import (
	"log"

	"github.com/Kaffyn/Vectora/src/cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		log.Fatal(err)
	}
}
