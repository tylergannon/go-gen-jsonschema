package main

import (
	"flag"
	"fmt"
)

func main() {
	// Define the --pretty flag
	pretty := flag.Bool("pretty", false, "Enable pretty output")

	// Parse the CLI arguments
	flag.Parse()

	// Use the flag value
	if *pretty {
		fmt.Println("Pretty output is enabled.")
	} else {
		fmt.Println("Pretty output is disabled.")
	}
}
