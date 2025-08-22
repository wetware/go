package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Child cell started!")
	fmt.Printf("Child name: %s\n", os.Getenv("WW_CHILD_NAME"))
	fmt.Printf("Child index: %s\n", os.Getenv("WW_CHILD_INDEX"))
	fmt.Printf("Parent PID: %d\n", os.Getppid())

	// Keep running for a bit to demonstrate
	fmt.Println("Child cell running... (press Ctrl+C to stop)")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Child cell shutting down...")
}
