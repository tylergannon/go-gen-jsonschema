package main

import (
	"log"

	"github.com/tylergannon/polytype/internal/builder"
)

func main() {
	if err := builder.Run(builder.BuilderArgs{TargetDir: ".", Pretty: true}); err != nil {
		log.Fatal(err)
	}
}
