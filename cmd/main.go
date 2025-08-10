package main

import (
	"os"

	"github.com/mbark/shmake"
)

func main() {
	shmake.RunStar(os.Args)
}
