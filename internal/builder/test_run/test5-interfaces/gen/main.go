package main

import (
	"log"

	"github.com/tylergannon/polytype/internal/builder"
)

func main() {
	err := builder.Run(builder.BuilderArgs{
		TargetDir:        ".",
		Pretty:           true,
		UnmarshalFormats: builder.UnmarshalFormatsBoth,
	})
	if err != nil {
		log.Fatal(err)
	}
}
