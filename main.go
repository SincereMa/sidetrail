package main

import (
	"fmt"
	"os"

	"github.com/SincereMa/cortex-sidemark/cmd/cortex"
)

func main() {
	if err := cortex.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
