package main

import (
	"context"
	"os"

	"github.com/mbark/shmake"
)

func main() {
	err := shmake.Run(context.Background(), os.Args)
	checkErr(err)
}

func checkErr(err error) {
	if err == nil {
		return
	}

	slog.Error(err.Error())
	os.Exit(1)
}
