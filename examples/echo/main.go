//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
)

//export echo
func echo() {
	if err := _echo(os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}

func _echo(dst io.Writer, src io.Reader) error {
	if dst == nil {
		return fmt.Errorf("destination writer is nil")
	}
	if src == nil {
		return fmt.Errorf("source reader is nil")
	}

	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func main() {
	stdin := flag.Bool("stdin", false, "read data from stdin")
	flag.Parse()

	if *stdin {
		if err := _echo(os.Stdout, os.Stdin); err != nil {
			slog.Error("failed to echo stdin",
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
