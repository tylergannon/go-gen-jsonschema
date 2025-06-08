package common

import (
	"io"
	"log"
)

func LogClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Printf("failed to close output file: %v", err)
	}
}
