// Package logfile implements opening of local log files.
package logfile

import (
	"io"
	"log"
	"os"
)

func Open(path string) io.Writer {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
