// Command yerk is the host multi-project manager CLI.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/salotz/yerk/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := cli.Execute(ctx, cli.DefaultIO(), os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "yerk: %v\n", err)
		os.Exit(1)
	}
}
