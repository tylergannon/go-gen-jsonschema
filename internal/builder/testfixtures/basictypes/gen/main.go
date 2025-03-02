package main

import (
	"log"

	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
)

func main() {
	err := builder.Run(builder.BuilderArgs{
		TargetDir:     ".",
		Pretty:        true,
		GenerateTests: false,
	})
	if err != nil {
		log.Fatal(err)
	}
}
