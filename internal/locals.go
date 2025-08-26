package internal

import (
	"fmt"
	"sync"

	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"
)

func InitialiseLocals(thread *starlark.Thread) (*CLI, *sync.WaitGroup) {
	wg := &sync.WaitGroup{}
	sindrCLI := &CLI{
		Command: &Command{
			Command: &cli.Command{},
		},
	}

	thread.SetLocal("cli", sindrCLI)
	thread.SetLocal("wg", wg)

	return sindrCLI, wg
}

func getSindrCLI(thread *starlark.Thread) (*CLI, error) {
	cliValue := thread.Local("cli")
	sindrCLI, ok := cliValue.(*CLI)
	if !ok {
		return nil, fmt.Errorf("expected cli to be set in thread local storage")
	}

	return sindrCLI, nil
}

func getWaitGroup(thread *starlark.Thread) (*sync.WaitGroup, error) {
	wgValue := thread.Local("wg")
	wg, ok := wgValue.(*sync.WaitGroup)
	if !ok {
		return nil, fmt.Errorf("expected wg to be set in thread local storage")
	}

	return wg, nil
}
