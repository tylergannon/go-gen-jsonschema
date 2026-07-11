package main

import (
	"log"

	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
)

func main() {
	if err := builder.Run(builder.BuilderArgs{TargetDir: ".", Pretty: true}); err != nil {
		log.Fatal(err)
	}
}
