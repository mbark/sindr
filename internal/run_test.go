package internal_test

import (
	"testing"

	"github.com/mbark/sindr/internal/sindrtest"
)

func TestPool(t *testing.T) {
	t.Run("creates pool and runs tasks concurrently", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    p = pool()
    
    def task1():
        shell('echo "task1 done" > task1.txt')
    
    def task2():
        shell('echo "task2 done" > task2.txt')
    
    p.run(task1)
    p.run(task2)
    p.wait()
    
    # Check files were created
    task1_result = shell('cat task1.txt')
    task2_result = shell('cat task2.txt')
    
    if str(task1_result) != 'task1 done':
        fail('expected "task1 done", got: ' + str(task1_result))
    if str(task2_result) != 'task2 done':
        fail('expected "task2 done", got: ' + str(task2_result))

cli(name="TestPool", usage="Test pool functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("pool waits for all tasks to complete", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    p = pool()
    
    def delayed_task():
        # Simulate some work with shell command
        shell('sleep 0.1')
        shell('echo "delayed task done" > delayed.txt')
    
    p.run(delayed_task)
    p.wait()  # This should wait for the delayed task
    
    result = shell('cat delayed.txt')
    if str(result) != 'delayed task done':
        fail('expected "delayed task done", got: ' + str(result))

cli(name="TestPool", usage="Test pool functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("multiple pools work independently", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    pool1 = pool()
    pool2 = pool()
    
    def pool1_task():
        shell('echo "pool1 task" > pool1.txt')
    
    def pool2_task():
        shell('echo "pool2 task" > pool2.txt')
    
    pool1.run(pool1_task)
    pool2.run(pool2_task)
    
    pool1.wait()
    pool2.wait()
    
    result1 = shell('cat pool1.txt')
    result2 = shell('cat pool2.txt')
    
    if str(result1) != 'pool1 task':
        fail('expected "pool1 task", got: ' + str(result1))
    if str(result2) != 'pool2 task':
        fail('expected "pool2 task", got: ' + str(result2))

cli(name="TestPool", usage="Test pool functionality")
command(name="test", action=test_action)
`)
	})
}

func TestAsync(t *testing.T) {
	t.Run("executes function asynchronously", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    def async_task():
        shell('echo "async task done" > async.txt')
    
    start(async_task)
    wait()  # Wait for async task to complete
    
    result = shell('cat async.txt')
    if str(result) != 'async task done':
        fail('expected "async task done", got: ' + str(result))

cli(name="TestAsync", usage="Test async functionality")
command(name="test", action=test_action)
`)
	})

	t.Run("multiple async tasks execute concurrently", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    def async1():
        shell('echo "async1 done" > async1.txt')
    
    def async2():
        shell('echo "async2 done" > async2.txt')
    
    start(async1)
    start(async2)
    wait()  # Wait for all async tasks
    
    result1 = shell('cat async1.txt')
    result2 = shell('cat async2.txt')
    
    if str(result1) != 'async1 done':
        fail('expected "async1 done", got: ' + str(result1))
    if str(result2) != 'async2 done':
        fail('expected "async2 done", got: ' + str(result2))

cli(name="TestAsync", usage="Test async functionality")
command(name="test", action=test_action)
`)
	})
}

func TestWait(t *testing.T) {
	t.Run("waits for async tasks to complete", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    def delayed_task():
        # Simulate work with shell command
        shell('sleep 0.1')
        shell('echo "completed" > wait_test.txt')
    
    start(delayed_task)
    
    # Before wait, task shouldn't be completed yet due to sleep
    wait()
    
    # After wait, task should be completed
    result = shell('cat wait_test.txt')
    if str(result) != 'completed':
        fail('expected "completed", got: ' + str(result))

cli(name="TestWait", usage="Test wait functionality")
command(name="test", action=test_action)
`)
	})
}

func TestRunTypeCreation(t *testing.T) {
	t.Run("pool function creates pool userdata", func(t *testing.T) {
		sindrtest.Test(t, `
def test_action(ctx):
    p = pool()
    
    # Verify pool has expected methods - in Starlark we can check attributes exist
    if not hasattr(p, 'run'):
        fail('expected pool to have run method')
    if not hasattr(p, 'wait'):
        fail('expected pool to have wait method')

cli(name="TestRunTypeCreation", usage="Test pool type creation")
command(name="test", action=test_action)
`)
	})
}
