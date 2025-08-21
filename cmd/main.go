package main

import (
	"context"
	"os"

	"github.com/mbark/shmake"
	"github.com/mbark/shmake/internal/logger"
)

func main() {
	err := shmake.Run(context.Background(), os.Args)
	if err != nil {
		logger.LogErr("error running shmake", err)
	}
}
