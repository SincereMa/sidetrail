package main

import (
	"fmt"
	"os"

	"github.com/SincereMa/cortex-sidemark/cmd/cortex"
)

// main is the entry point. It delegates to the cortex command
// and exits non-zero on error. The command package is the only
// place the CLI surface is wired; main.go stays trivially short.
func main() {
	if err := cortex.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
