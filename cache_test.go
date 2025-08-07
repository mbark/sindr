package shmake_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mbark/shmake"
)

func withMainLua(t *testing.T, dir string, contents string) {
	err := os.RemoveAll(filepath.Join(dir, "main.lua"))
	require.NoError(t, err)

	f, err := os.Create(filepath.Join(dir, "main.lua"))
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

	cacheDir := filepath.Join(dir, "cache")
	logFile, err := os.Create(filepath.Join(dir, "logs"))
	require.NoError(t, err)
	t.Cleanup(func() {
		err := logFile.Close()
		require.NoError(t, err)
	})

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

func TestStore(t *testing.T) {
	dir := t.TempDir()

	cacheDir := filepath.Join(dir, "cache")
	logFile, err := os.Create(filepath.Join(dir, "logs"))
	require.NoError(t, err)
	t.Cleanup(func() {
		err := logFile.Close()
		require.NoError(t, err)
	})

	err = os.Chdir(dir)
	require.NoError(t, err)

	t.Run("store version successfully", func(t *testing.T) {
		withMainLua(t, dir, `
local shmake = require("shmake.main")

local cli = shmake.command('TestStore')

cli:command('test'):action(function()
	shmake.store({name='test-key', version='v1.0.0'})
	
	-- Verify it was stored by checking with get_version
	local stored = shmake.get_version('test-key')
	if stored ~= 'v1.0.0' then
		error('expected stored version to be v1.0.0, got: ' .. tostring(stored))
	end
end)

cli:run()
`)
		r, err := shmake.NewRuntime(cacheDir, logFile)
		require.NoError(t, err)
		r.Args = []string{"TestStore", "test"}

		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("store with int version", func(t *testing.T) {
		withMainLua(t, dir, `
local shmake = require("shmake.main")

local cli = shmake.command('TestStore')

cli:command('test'):action(function()
	shmake.store({name='test-int', int_version=42})
	
	-- Verify it was stored by checking with get_version
	local stored = shmake.get_version('test-int')
	if stored ~= '42' then
		error('expected stored version to be 42, got: ' .. tostring(stored))
	end
end)

cli:run()
`)
		r, err := shmake.NewRuntime(cacheDir, logFile)
		require.NoError(t, err)
		r.Args = []string{"TestStore", "test"}

		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestWithVersion(t *testing.T) {
	dir := t.TempDir()

	cacheDir := filepath.Join(dir, "cache")
	logFile, err := os.Create(filepath.Join(dir, "logs"))
	require.NoError(t, err)
	t.Cleanup(func() {
		err := logFile.Close()
		require.NoError(t, err)
	})

	err = os.Chdir(dir)
	require.NoError(t, err)

	t.Run("runs function when version differs", func(t *testing.T) {
		withMainLua(t, dir, `
local shmake = require("shmake.main")

local cli = shmake.command('TestWithVersion')
local executed = false

cli:command('test'):action(function()
	local ran = shmake.with_version({name='test-version', version='v2.0.0'}, function()
		executed = true
	end)
	
	if not ran then
		error('expected with_version to return true when function runs')
	end
	
	if not executed then
		error('expected function to be executed')
	end
end)

cli:run()
`)
		r, err := shmake.NewRuntime(cacheDir, logFile)
		require.NoError(t, err)
		r.Args = []string{"TestWithVersion", "test"}

		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("skips function when version matches", func(t *testing.T) {
		withMainLua(t, dir, `
local shmake = require("shmake.main")

local cli = shmake.command('TestWithVersion')
local executed = false

cli:command('test'):action(function()
	-- First store a version
	shmake.store({name='skip-test', version='v1.5.0'})
	
	-- Then try to run with_version with same version
	local ran = shmake.with_version({name='skip-test', version='v1.5.0'}, function()
		executed = true
	end)
	
	if ran then
		error('expected with_version to return false when versions match')
	end
	
	if executed then
		error('expected function to be skipped')
	end
end)

cli:run()
`)
		r, err := shmake.NewRuntime(cacheDir, logFile)
		require.NoError(t, err)
		r.Args = []string{"TestWithVersion", "test"}

		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("runs function with int version", func(t *testing.T) {
		withMainLua(t, dir, `
local shmake = require("shmake.main")

local cli = shmake.command('TestWithVersion')
local executed = false

cli:command('test'):action(function()
	local ran = shmake.with_version({name='int-version', int_version=123}, function()
		executed = true
	end)
	
	if not ran then
		error('expected with_version to return true when function runs')
	end
	
	if not executed then
		error('expected function to be executed')
	end
	
	-- Verify version was stored
	local stored = shmake.get_version('int-version')
	if stored ~= '123' then
		error('expected stored version to be 123, got: ' .. tostring(stored))
	end
end)

cli:run()
`)
		r, err := shmake.NewRuntime(cacheDir, logFile)
		require.NoError(t, err)
		r.Args = []string{"TestWithVersion", "test"}

		shmake.RunWithRuntime(t.Context(), r)
	})
}
