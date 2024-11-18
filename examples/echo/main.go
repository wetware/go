//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"bufio"
	"flag"
	"io"
	"log/slog"
	"os"
)

func main() {
	stdin := flag.Bool("stdin", false, "read data from stdin")
	flag.Parse()

	if *stdin {
		if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
			slog.Error("failed echo stdin",
				"reason", err)
			os.Exit(1)
		}

	} else {
		w := bufio.NewWriter(os.Stdout)
		for _, arg := range os.Args[1:] {
			w.WriteString(arg)
			w.WriteString(" ")
			if err := w.Flush(); err != nil {
				slog.Error("failed to flush argument to stdout",
					"reason", err)
				os.Exit(1)
			}
		}
	}
}

//export echo
func echo() {
	if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
