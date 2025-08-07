package shmake_test

import (
	"testing"

	"github.com/mbark/shmake"
	"github.com/stretchr/testify/require"
)

func TestDiff(t *testing.T) {
	t.Run("with no diff expected", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestDiff')

cli:command('test'):action(function()
	if not shmake.diff({name='version', version='1'}) then
		error('unexpected diff')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("with diff expected", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestDiff')

cli:command('test'):action(function()
	if shmake.diff({name='version', version='1'}) then
		error('expected no diff')
	end
end)

cli:run()
`)

		err := r.Cache.StoreVersion("version", "1")
		require.NoError(t, err)

		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestStore(t *testing.T) {
	t.Run("store version successfully", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
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
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("store with int version", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
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
		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestWithVersion(t *testing.T) {
	t.Run("runs function when version differs", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
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
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("skips function when version matches", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
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
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("runs function with int version", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
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
		shmake.RunWithRuntime(t.Context(), r)
	})
}
