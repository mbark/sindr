package shmake_test

import (
	"testing"

	"github.com/mbark/shmake"
)

func TestPool(t *testing.T) {
	t.Run("creates pool and runs tasks concurrently", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestPool')

cli:command('test'):action(function()
	local p = shmake.pool()
	local counter = 0
	
	p:run(function()
		files.write('task1.txt', 'task1 done')
		counter = counter + 1
	end)
	
	p:run(function()
		files.write('task2.txt', 'task2 done')
		counter = counter + 1
	end)
	
	p:wait()
	
	-- Check files were created
	local task1 = shmake.shell('cat task1.txt')
	local task2 = shmake.shell('cat task2.txt')
	
	if task1 ~= 'task1 done' then
		error('expected task1 done, got: ' .. tostring(task1))
	end
	if task2 ~= 'task2 done' then
		error('expected task2 done, got: ' .. tostring(task2))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("pool waits for all tasks to complete", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestPool')

cli:command('test'):action(function()
	local p = shmake.pool()
	
	p:run(function()
		-- Simulate some work with shell command
		shmake.shell('sleep 0.1')
		files.write('delayed.txt', 'delayed task done')
	end)
	
	p:wait()  -- This should wait for the delayed task
	
	local result = shmake.shell('cat delayed.txt')
	if result ~= 'delayed task done' then
		error('expected delayed task done, got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("multiple pools work independently", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestPool')

cli:command('test'):action(function()
	local p1 = shmake.pool()
	local p2 = shmake.pool()
	
	p1:run(function()
		files.write('pool1.txt', 'pool1 task')
	end)
	
	p2:run(function()
		files.write('pool2.txt', 'pool2 task')
	end)
	
	p1:wait()
	p2:wait()
	
	local result1 = shmake.shell('cat pool1.txt')
	local result2 = shmake.shell('cat pool2.txt')
	
	if result1 ~= 'pool1 task' then
		error('expected pool1 task, got: ' .. tostring(result1))
	end
	if result2 ~= 'pool2 task' then
		error('expected pool2 task, got: ' .. tostring(result2))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestAsync(t *testing.T) {
	t.Run("executes function asynchronously", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestAsync')

cli:command('test'):action(function()
	shmake.async(function()
		files.write('async.txt', 'async task done')
	end)
	
	shmake.wait()  -- Wait for async task to complete
	
	local result = shmake.shell('cat async.txt')
	if result ~= 'async task done' then
		error('expected async task done, got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})

	t.Run("multiple async tasks execute concurrently", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestAsync')

cli:command('test'):action(function()
	shmake.async(function()
		files.write('async1.txt', 'async1 done')
	end)
	
	shmake.async(function()
		files.write('async2.txt', 'async2 done')
	end)
	
	shmake.wait()  -- Wait for all async tasks
	
	local result1 = shmake.shell('cat async1.txt')
	local result2 = shmake.shell('cat async2.txt')
	
	if result1 ~= 'async1 done' then
		error('expected async1 done, got: ' .. tostring(result1))
	end
	if result2 ~= 'async2 done' then
		error('expected async2 done, got: ' .. tostring(result2))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestWait(t *testing.T) {
	t.Run("waits for async tasks to complete", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")
local files = require("shmake.files")

local cli = shmake.command('TestWait')

cli:command('test'):action(function()
	local completed = false
	
	shmake.async(function()
		-- Simulate work with shell command
		shmake.shell('sleep 0.1')
		files.write('wait_test.txt', 'completed')
		completed = true
	end)
	
	-- Before wait, task shouldn't be completed yet due to sleep
	shmake.wait()
	
	-- After wait, task should be completed
	local result = shmake.shell('cat wait_test.txt')
	if result ~= 'completed' then
		error('expected completed, got: ' .. tostring(result))
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}

func TestRunTypeCreation(t *testing.T) {
	t.Run("pool function creates pool userdata", func(t *testing.T) {
		r := setupRuntime(t)
		withMainLua(t, `
local shmake = require("shmake.main")

local cli = shmake.command('TestRunTypeCreation')

cli:command('test'):action(function()
	local p = shmake.pool()
	
	-- Verify pool has expected methods
	if type(p.run) ~= 'function' then
		error('expected pool to have run method')
	end
	if type(p.wait) ~= 'function' then
		error('expected pool to have wait method')
	end
end)

cli:run()
`)
		shmake.RunWithRuntime(t.Context(), r)
	})
}
