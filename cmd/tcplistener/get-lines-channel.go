package main

import (
	"io"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
		var line string
		var err error
		for err != io.EOF {
			out := make([]byte, 8, 8)
			n, err := f.Read(out)
			if err == io.EOF {
				f.Close()
				break
			}
			split := strings.Split(string(out[:n]), "\n")
			if len(split) != 1 {
				line += split[0]
				lines <- line
				line = split[1]
			} else {
				line += split[0]
			}
		}
		if line != "" {
			lines <- line
		}
		close(lines)
	}()
	return lines
}
