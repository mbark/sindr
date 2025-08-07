package shmake_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mbark/shmake"
)

func withMainLua(t *testing.T, dir string, contents string) {
	err := os.RemoveAll(dir + "/main.lua")
	require.NoError(t, err)

	f, err := os.Create(dir + "/main.lua")
	require.NoError(t, err)

	t.Cleanup(func() {
		err := f.Close()
		require.NoError(t, err)
	})

	_, err = f.WriteString(contents)
	require.NoError(t, err)

	t.Log("=== main.lua ===")
	for i, line := range strings.Split(contents, "\n") {
		t.Logf("%3d: %s", i+1, line)
	}
	t.Log()
}

func TestDiff(t *testing.T) {
	dir := t.TempDir()

	cacheDir := dir + "/cache"
	logFile, err := os.Create(dir + "/logs")
	require.NoError(t, err)
	defer logFile.Close()

	err = os.Chdir(dir)
	require.NoError(t, err)

	t.Run("with no diff expected", func(t *testing.T) {
		withMainLua(t, dir, `
local shmake = require("shmake.main")

local cli = shmake.command('TestDiff')

cli:command('test'):action(function()
	if not shmake.diff({name='version', version='1'}) then
		error('unexpected diff')
	end
end)

cli:run()
`)
		r, err := shmake.NewRuntime(cacheDir, logFile)
		require.NoError(t, err)
		r.Args = []string{"TestDiff", "test"}

		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("with diff expected", func(t *testing.T) {
		withMainLua(t, dir, `
local shmake = require("shmake.main")

local cli = shmake.command('TestDiff')

cli:command('test'):action(function()
	if shmake.diff({name='version', version='1'}) then
		error('expected no diff')
	end
end)

cli:run()
`)
		r, err := shmake.NewRuntime(cacheDir, logFile)
		require.NoError(t, err)
		r.Args = []string{"TestDiff", "test"}
		err = r.Cache.StoreVersion("version", "1")
		require.NoError(t, err)

		shmake.RunWithRuntime(t.Context(), r)
	})
}
