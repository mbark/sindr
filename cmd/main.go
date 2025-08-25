package main

import (
	"context"
	"os"

	"github.com/mbark/sindr"
	"github.com/mbark/sindr/internal/logger"
)

func main() {
	err := sindr.Run(context.Background(), os.Args)
	if err != nil {
		logger.LogErr("error running sindr", err)
	}
}
