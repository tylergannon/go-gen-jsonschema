package main

import (
	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
	"log"
)

func main() {
	if err := builder.Run(builder.BuilderArgs{TargetDir: ".", Pretty: true}); err != nil {
		log.Fatal(err)
	}
}
