package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	lua "github.com/yuin/gopher-lua"
)

func getGitModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"head": head(runtime),
		},
	}
}

func findGitDir(start string) (string, error) {
	basePath := "/"
	targetPath := start

	for {
		if rel, _ := filepath.Rel(basePath, targetPath); rel == "." {
			break
		}

		path := fmt.Sprintf("%v/%v", targetPath, git.GitDirName)
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				targetPath += "/.."
				continue
			}

			panic(err)
		}

		return filepath.Abs(path)
	}

	return "", errors.New("no git directory found")
}

func head(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		dir, err := findGitDir(".")
		if err != nil {
			panic(err)
		}

		repo, err := git.PlainOpen(dir)
		if err != nil {
			panic(err)
		}

		head, err := repo.Head()
		if err != nil {
			panic(err)
		}

		L.Push(lua.LString(head.Hash().String()))
		return 1
	}
}
