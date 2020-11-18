package main

import (
	"path"

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

func head(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		dir, err := findPathUpdwards(git.GitDirName)
		if err != nil {
			panic(err)
		}

		repo, err := git.PlainOpen(path.Join(dir, git.GitDirName))
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
