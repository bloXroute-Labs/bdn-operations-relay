package main

import (
	"fmt"
	"os"
)

func main() {
	if err := relayCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}
