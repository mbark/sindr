package shmake_test

import (
	"testing"

	"github.com/mbark/shmake"
)

func TestShell(t *testing.T) {
	t.Run("executes basic shell command", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestShell')

cli:command('test'):action(function()
	local result = shmake.shell('echo "hello world"')
	if result ~= 'hello world' then
		error('expected "hello world", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("captures command output", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestShell')

cli:command('test'):action(function()
	local result = shmake.shell('printf "line1\nline2\nline3"')
	local expected = 'line1\nline2\nline3'
	if result ~= expected then
		error('expected: ' .. expected .. ', got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("works with shell variables", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestShell')

cli:command('test'):action(function()
	local result = shmake.shell('VAR="test value" && echo $VAR')
	if result ~= 'test value' then
		error('expected "test value", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("handles command with options prefix", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestShell')

cli:command('test'):action(function()
	local result = shmake.shell('echo "prefixed output"', {prefix = 'TEST'})
	if result ~= 'prefixed output' then
		error('expected "prefixed output", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("trims whitespace from output", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestShell')

cli:command('test'):action(function()
	local result = shmake.shell('echo "  content with spaces  "')
	if result ~= 'content with spaces' then
		error('expected "  content with spaces  ", got: ' .. tostring(result))
	end
	
	-- Test trailing newline is trimmed
	local result2 = shmake.shell('printf "no newline here"')
	if result2 ~= 'no newline here' then
		error('expected "no newline here", got: ' .. tostring(result2))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("handles empty command output", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestShell')

cli:command('test'):action(function()
	local result = shmake.shell('true')  -- command that produces no output
	if result ~= '' then
		error('expected empty string, got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("works with complex commands", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestShell')

cli:command('test'):action(function()
	-- Create a test file and read it back
	shmake.shell('echo "test content" > test.txt')
	local result = shmake.shell('cat test.txt')
	if result ~= 'test content' then
		error('expected "test content", got: ' .. tostring(result))
	end
	
	-- Clean up
	shmake.shell('rm test.txt')
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}
