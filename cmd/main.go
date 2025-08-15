package main

import (
	"context"
	"os"

	"github.com/mbark/shmake"
)

func main() {
	shmake.Run(context.Background(), os.Args)
}
