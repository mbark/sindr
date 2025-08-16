package internal_test

import (
	"testing"
)

func TestPool(t *testing.T) {
	t.Run("creates pool and runs tasks concurrently", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    pool = shmake.pool()
    
    def task1():
        shmake.shell('echo "task1 done" > task1.txt')
    
    def task2():
        shmake.shell('echo "task2 done" > task2.txt')
    
    pool.run(task1)
    pool.run(task2)
    pool.wait()
    
    # Check files were created
    task1_result = shmake.shell('cat task1.txt')
    task2_result = shmake.shell('cat task2.txt')
    
    if str(task1_result) != 'task1 done':
        fail('expected "task1 done", got: ' + str(task1_result))
    if str(task2_result) != 'task2 done':
        fail('expected "task2 done", got: ' + str(task2_result))

shmake.cli(name="TestPool", usage="Test pool functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("pool waits for all tasks to complete", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    pool = shmake.pool()
    
    def delayed_task():
        # Simulate some work with shell command
        shmake.shell('sleep 0.1')
        shmake.shell('echo "delayed task done" > delayed.txt')
    
    pool.run(delayed_task)
    pool.wait()  # This should wait for the delayed task
    
    result = shmake.shell('cat delayed.txt')
    if str(result) != 'delayed task done':
        fail('expected "delayed task done", got: ' + str(result))

shmake.cli(name="TestPool", usage="Test pool functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("multiple pools work independently", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    pool1 = shmake.pool()
    pool2 = shmake.pool()
    
    def pool1_task():
        shmake.shell('echo "pool1 task" > pool1.txt')
    
    def pool2_task():
        shmake.shell('echo "pool2 task" > pool2.txt')
    
    pool1.run(pool1_task)
    pool2.run(pool2_task)
    
    pool1.wait()
    pool2.wait()
    
    result1 = shmake.shell('cat pool1.txt')
    result2 = shmake.shell('cat pool2.txt')
    
    if str(result1) != 'pool1 task':
        fail('expected "pool1 task", got: ' + str(result1))
    if str(result2) != 'pool2 task':
        fail('expected "pool2 task", got: ' + str(result2))

shmake.cli(name="TestPool", usage="Test pool functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}

func TestAsync(t *testing.T) {
	t.Run("executes function asynchronously", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    def async_task():
        shmake.shell('echo "async task done" > async.txt')
    
    shmake.start(async_task)
    shmake.wait()  # Wait for async task to complete
    
    result = shmake.shell('cat async.txt')
    if str(result) != 'async task done':
        fail('expected "async task done", got: ' + str(result))

shmake.cli(name="TestAsync", usage="Test async functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})

	t.Run("multiple async tasks execute concurrently", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    def async1():
        shmake.shell('echo "async1 done" > async1.txt')
    
    def async2():
        shmake.shell('echo "async2 done" > async2.txt')
    
    shmake.start(async1)
    shmake.start(async2)
    shmake.wait()  # Wait for all async tasks
    
    result1 = shmake.shell('cat async1.txt')
    result2 = shmake.shell('cat async2.txt')
    
    if str(result1) != 'async1 done':
        fail('expected "async1 done", got: ' + str(result1))
    if str(result2) != 'async2 done':
        fail('expected "async2 done", got: ' + str(result2))

shmake.cli(name="TestAsync", usage="Test async functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}

func TestWait(t *testing.T) {
	t.Run("waits for async tasks to complete", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    def delayed_task():
        # Simulate work with shell command
        shmake.shell('sleep 0.1')
        shmake.shell('echo "completed" > wait_test.txt')
    
    shmake.start(delayed_task)
    
    # Before wait, task shouldn't be completed yet due to sleep
    shmake.wait()
    
    # After wait, task should be completed
    result = shmake.shell('cat wait_test.txt')
    if str(result) != 'completed':
        fail('expected "completed", got: ' + str(result))

shmake.cli(name="TestWait", usage="Test wait functionality")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}

func TestRunTypeCreation(t *testing.T) {
	t.Run("pool function creates pool userdata", func(t *testing.T) {
		run := setupStarlarkRuntime(t)
		withMainStar(t, `
def test_action(ctx):
    pool = shmake.pool()
    
    # Verify pool has expected methods - in Starlark we can check attributes exist
    if not hasattr(pool, 'run'):
        fail('expected pool to have run method')
    if not hasattr(pool, 'wait'):
        fail('expected pool to have wait method')

shmake.cli(name="TestRunTypeCreation", usage="Test pool type creation")
shmake.command(name="test", action=test_action)
`)
		run()
	})
}
