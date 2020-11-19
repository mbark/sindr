package main

import (
	"path"

	"github.com/go-git/go-git/v5"
	lua "github.com/yuin/gopher-lua"
)

func getGitModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]ModuleFunction{
			"head": head,
		},
	}
}

func head(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	dir, err := findPathUpdwards(git.GitDirName)
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(path.Join(dir, git.GitDirName))
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	return []lua.LValue{lua.LString(head.Hash().String())}, nil
}
