package shmake_test

import (
	"testing"

	"github.com/mbark/shmake"
)

func TestFileWrite(t *testing.T) {
	t.Run("writes content to file", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileWrite')

cli:command('test'):action(function()
	files.write('test.txt', 'hello world')
	local result = shmake.shell('cat test.txt')
	if result ~= 'hello world' then
		error('expected "hello world", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileWrite')

cli:command('test'):action(function()
	files.write('test.txt', 'first content')
	files.write('test.txt', 'second content')
	local result = shmake.shell('cat test.txt')
	if result ~= 'second content' then
		error('expected "second content", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestFileDelete(t *testing.T) {
	t.Run("deletes single file with string", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileDelete')

cli:command('test'):action(function()
	files.write('test.txt', 'content')
	files.delete('test.txt')
	local result = shmake.shell('ls test.txt 2>/dev/null || echo "not found"')
	if result ~= 'not found' then
		error('expected file to be deleted, but it still exists')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("deletes multiple files with glob", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileDelete')

cli:command('test'):action(function()
	files.write('test1.txt', 'content1')
	files.write('test2.txt', 'content2')
	files.write('other.log', 'content3')
	files.delete('test*.txt')
	
	local result1 = shmake.shell('ls test1.txt 2>/dev/null || echo "not found"')
	local result2 = shmake.shell('ls test2.txt 2>/dev/null || echo "not found"')
	local result3 = shmake.shell('ls other.log 2>/dev/null || echo "not found"')
	
	if result1 ~= 'not found' then
		error('expected test1.txt to be deleted')
	end
	if result2 ~= 'not found' then
		error('expected test2.txt to be deleted')
	end
	if result3 == 'not found' then
		error('expected other.log to still exist')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("deletes directory", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileDelete')

cli:command('test'):action(function()
	files.mkdir('testdir')
	files.write('testdir/file.txt', 'content')
	files.delete('testdir')
	
	local result = shmake.shell('ls testdir 2>/dev/null || echo "not found"')
	if result ~= 'not found' then
		error('expected directory to be deleted')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestFileCopy(t *testing.T) {
	t.Run("copies file content", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileCopy')

cli:command('test'):action(function()
	files.write('source.txt', 'original content')
	files.copy({from = 'source.txt', to = 'dest.txt'})
	
	local result = shmake.shell('cat dest.txt')
	if result ~= 'original content' then
		error('expected "original content", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("overwrites destination file", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileCopy')

cli:command('test'):action(function()
	files.write('source.txt', 'new content')
	files.write('dest.txt', 'old content')
	files.copy({from = 'source.txt', to = 'dest.txt'})
	
	local result = shmake.shell('cat dest.txt')
	if result ~= 'new content' then
		error('expected "new content", got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestFileMkdir(t *testing.T) {
	t.Run("creates single directory", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileMkdir')

cli:command('test'):action(function()
	files.mkdir('testdir')
	local result = shmake.shell('test -d testdir && echo "exists" || echo "not found"')
	if result ~= 'exists' then
		error('expected directory to be created')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("creates nested directories with all option", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileMkdir')

cli:command('test'):action(function()
	files.mkdir('parent/child/nested', {all = true})
	local result = shmake.shell('test -d parent/child/nested && echo "exists" || echo "not found"')
	if result ~= 'exists' then
		error('expected nested directory to be created')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("ignores existing directory", func(t *testing.T) {
		r := setupRuntime(t)
		//nolint:dupword // Intentionally testing mkdir on existing directory
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileMkdir')

cli:command('test'):action(function()
	files.mkdir('testdir')
	files.mkdir('testdir')
	local result = shmake.shell('test -d testdir && echo "exists" || echo "not found"')
	if result ~= 'exists' then
		error('expected directory to still exist after second mkdir')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestFileChdir(t *testing.T) {
	t.Run("changes and restores directory", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileChdir')

cli:command('test'):action(function()
	local original = shmake.shell('pwd')
	files.mkdir('subdir')
	files.chdir('subdir')
	local new_dir = shmake.shell('pwd')
	files.popdir()
	local restored = shmake.shell('pwd')
	
	if new_dir == original then
		error('directory did not change')
	end
	if restored ~= original then
		error('directory was not restored, expected: ' .. original .. ', got: ' .. restored)
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestFileTimestamp(t *testing.T) {
	t.Run("returns newest timestamp", func(t *testing.T) {
		r := setupRuntime(t)

		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileTimestamp')

cli:command('test'):action(function()
	-- Create files with different timestamps using the files module
	files.write('file1.txt', 'content1')
	shmake.shell('sleep 0.1') -- Ensure different timestamps
	files.write('file2.txt', 'content2')
	
	local newest = files.newest_ts('file*.txt')
	local oldest = files.oldest_ts('file*.txt')
	
	if newest < oldest then
		error('newest timestamp should be >= oldest, newest=' .. tostring(newest) .. ', oldest=' .. tostring(oldest))
	end
	if newest == 0 then
		error('newest timestamp should not be 0')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("returns 0 for no matching files", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestFileTimestamp')

cli:command('test'):action(function()
	local timestamp = files.newest_ts('nonexistent*.txt')
	if timestamp ~= 0 then
		error('expected timestamp to be 0 for non-existent files, got: ' .. tostring(timestamp))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}
