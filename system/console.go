package system

import (
	"context"
	"fmt"
	"io"
)

// ConsoleConfig follows the same pattern as system.IPFSConfig and system.ExecConfig
type ConsoleConfig struct {
	Writer io.Writer
}

// New creates a new Console client from the config
func (c ConsoleConfig) New() Console {
	if c.Writer == nil {
		return Console{ /* null capability */ }
	}

	return Console_ServerToClient(c)
}

func (c ConsoleConfig) Println(ctx context.Context, call Console_println) error {
	// Get the output data from the call
	output, err := call.Args().Output()
	if err != nil {
		return err
	}

	// Write to the writer with a newline
	bytesWritten, err := fmt.Fprintln(c.Writer, output)
	if err != nil {
		return err
	}

	// Set the result (number of bytes written)
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	results.SetN(uint32(bytesWritten))

	return nil
}
