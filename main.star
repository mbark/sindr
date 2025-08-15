# Global variables for template examples
project_name = "shmake-demo"
version = "1.0.0"
build_dir = "./build"

# Configure CLI metadata
shmake.cli(
    name = "shmake",
    usage = "Comprehensive demo of shmake features and functions"
)

# ============================================================================
# BASIC SHELL COMMAND EXAMPLE
# ============================================================================
def demo_shell(ctx):
    """Demonstrates shmake.shell() with different options"""
    print("=== Shell Command Demo ===")
    
    # Basic shell command
    result = shmake.shell('echo "Hello from shmake!"')
    print("Shell output:", result.stdout)
    print("Command success:", result.success)
    print("Exit code:", result.exit_code)
    
    # Shell command with prefix
    shmake.shell('echo "This has a prefix"', prefix='[DEMO]')
    
    # Multiple commands
    shmake.shell('echo "Command 1" && echo "Command 2"')

shmake.command(
    name = "shell",
    help = "Demonstrate shell command execution",
    action = demo_shell
)

# ============================================================================
# ASYNC OPERATIONS EXAMPLE
# ============================================================================
def demo_async(ctx):
    """Demonstrates async operations with shmake.start() and shmake.wait()"""
    print("=== Async Operations Demo ===")
    
    # Start multiple operations concurrently
    shmake.start(lambda: shmake.shell('sleep 1 && echo "Task 1 completed"', prefix='[TASK1]'))
    shmake.start(lambda: shmake.shell('sleep 2 && echo "Task 2 completed"', prefix='[TASK2]'))
    shmake.start(lambda: shmake.shell('sleep 1.5 && echo "Task 3 completed"', prefix='[TASK3]'))
    
    print("Started 3 async tasks, waiting for completion...")
    shmake.wait()
    print("All async tasks completed!")

shmake.command(
    name = "async",
    help = "Demonstrate async operations with start/wait",
    action = demo_async
)

# ============================================================================
# POOL OPERATIONS EXAMPLE
# ============================================================================
def demo_pool(ctx):
    """Demonstrates pool-based concurrent operations"""
    print("=== Pool Operations Demo ===")
    
    pool = shmake.pool()
    
    # Add tasks to the pool
    pool.run(lambda: shmake.shell('echo "Pool task 1" && sleep 1', prefix='[POOL1]'))
    pool.run(lambda: shmake.shell('echo "Pool task 2" && sleep 1', prefix='[POOL2]'))
    pool.run(lambda: shmake.shell('echo "Pool task 3" && sleep 1', prefix='[POOL3]'))
    
    print("Pool tasks started, waiting...")
    pool.wait()
    print("Pool tasks completed!")

shmake.command(
    name = "pool",
    help = "Demonstrate pool-based concurrent operations",
    action = demo_pool
)

# ============================================================================
# STRING TEMPLATING EXAMPLE
# ============================================================================
def demo_templates(ctx):
    """Demonstrates string templating with global and local variables"""
    print("=== String Templating Demo ===")
    
    # Using global variables in template
    template1 = shmake.string('''
Project: {{.project_name}}
Version: {{.version}}
Build Directory: {{.build_dir}}
''', project_name=project_name, version=version, build_dir=build_dir)
    print("Template with global variables:")
    print(template1)
    
    # Using local variables in template
    template2 = shmake.string('''
Environment: {{.env}}
Database: {{.db_host}}:{{.db_port}}
Debug Mode: {{.debug}}
''', env='production', db_host='localhost', db_port=5432, debug=False)
    
    print("Template with local variables:")
    print(template2)
    
    # Complex template with conditionals and loops
    template3 = shmake.string('''
{{if .enable_feature}}Feature is enabled!{{else}}Feature is disabled{{end}}
Services: {{range .services}}
  - {{.}}{{end}}
''', enable_feature=True, services=['web', 'api', 'worker'])
    
    print("Complex template:")
    print(template3)

shmake.command(
    name = "templates",
    help = "Demonstrate string templating features",
    action = demo_templates
)

# ============================================================================
# CACHE INSTANCE EXAMPLE
# ============================================================================
def demo_versioning(ctx):
    """Demonstrates version storage and caching with cache instances"""
    print("=== Versioning and Caching Demo ===")
    
    # Create a cache instance
    c = cache()
    
    # Set a version
    c.set_version(name="build", version="1.2.3")
    current_version = c.get_version("build")
    print("Current build version:", current_version)
    
    # Use with_version for expensive operations
    def expensive_operation():
        print("Running expensive build operation...")
        shmake.shell('sleep 2 && echo "Build completed!"')
        return "build-artifacts.tar.gz"
    
    # This will run the function only if version differs
    ran = c.with_version(expensive_operation, name="build", version="1.2.3")
    if ran:
        print("Build operation executed")
    else:
        print("Build skipped - version unchanged")
    
    # Running again with same version will skip execution
    print("Running again (should skip):")
    ran2 = c.with_version(expensive_operation, name="build", version="1.2.3")
    if ran2:
        print("Build operation executed again")
    else:
        print("Build skipped - version unchanged")
    
    # Running with different version will execute
    print("Running with new version:")
    ran3 = c.with_version(expensive_operation, name="build", version="1.2.4")
    if ran3:
        print("Build operation executed with new version")
    else:
        print("Build skipped")

shmake.command(
    name = "versioning",
    help = "Demonstrate versioning and caching with cache instances",
    action = demo_versioning
)

# ============================================================================
# CACHE DIFF EXAMPLE
# ============================================================================
def demo_diff(ctx):
    """Demonstrates version comparison with cache.diff()"""
    print("=== Version Diff Demo ===")
    
    # Create a cache instance
    c = cache()
    
    # Check if version differs (should return True for first time)
    config_version = "v1.0.0"
    has_diff = c.diff(name="config", version=config_version)
    print("Config version differs:", has_diff)
    
    # Set the version
    c.set_version(name="config", version=config_version)
    
    # Check again (should return False now)
    has_diff2 = c.diff(name="config", version=config_version)
    print("Config version differs after setting:", has_diff2)
    
    # Check with different version (should return True)
    new_version = "v1.1.0"
    has_diff3 = c.diff(name="config", version=new_version)
    print("New config version differs:", has_diff3)
    
    # Demonstrate with integer versions
    build_num = 42
    has_diff4 = c.diff(name="build_number", version=build_num)
    print("Build number differs:", has_diff4)

shmake.command(
    name = "diff",
    help = "Demonstrate version comparison with cache.diff()",
    action = demo_diff
)

# ============================================================================
# COMPREHENSIVE BUILD EXAMPLE
# ============================================================================
def comprehensive_build(ctx):
    """A comprehensive example combining multiple features"""
    print("=== Comprehensive Build Demo ===")
    
    # Get build configuration from arguments and flags
    target = ctx.args.target if ctx.args.target else "all"
    verbose = ctx.flags.verbose
    parallel = ctx.flags.parallel
    
    if verbose:
        print("Building target:", target)
        print("Parallel build:", parallel)
    
    # Create cache instance and set build version using shell command to get timestamp
    c = cache()
    timestamp_result = shmake.shell('date +%s')
    build_version = "build-" + timestamp_result.stdout.strip()
    c.set_version(name="current_build", version=build_version)
    
    # Create build script using templates
    build_script = shmake.string('''#!/bin/bash
echo "Building {{.project_name}} v{{.version}}"
echo "Target: {{.target}}"
echo "Build ID: {{.build_id}}"
mkdir -p {{.build_dir}}
echo "Build completed at $(date)" > {{.build_dir}}/build.log
''', target=target, build_id=build_version)
    
    if verbose:
        print("Generated build script:")
        print(build_script)
    
    # Execute build based on flags
    if parallel:
        print("Running parallel build...")
        pool = shmake.pool()
        pool.run(lambda: shmake.shell('echo "Compiling frontend..." && sleep 1', prefix='[FRONTEND]'))
        pool.run(lambda: shmake.shell('echo "Compiling backend..." && sleep 2', prefix='[BACKEND]'))
        pool.run(lambda: shmake.shell('echo "Running tests..." && sleep 1.5', prefix='[TESTS]'))
        pool.wait()
    else:
        print("Running sequential build...")
        shmake.shell('echo "Compiling frontend..."', prefix='[BUILD]')
        shmake.shell('echo "Compiling backend..."', prefix='[BUILD]')
        shmake.shell('echo "Running tests..."', prefix='[BUILD]')
    
    # Final status
    shmake.shell(shmake.string('echo "Build {{.build_id}} completed successfully!"', build_id=build_version))

shmake.command(
    name = "build",
    help = "Comprehensive build example with args and flags",
    action = comprehensive_build,
    args = ["target"],
    flags = {
        "verbose": {
            "type": "bool",
            "default": False,
            "help": "Enable verbose output"
        },
        "parallel": {
            "type": "bool", 
            "default": False,
            "help": "Enable parallel build"
        }
    }
)

# ============================================================================
# SUB-COMMAND EXAMPLES
# ============================================================================
def deploy_staging(ctx):
    print("Deploying to staging environment...")
    shmake.shell('echo "Staging deployment started"', prefix='[STAGING]')

def deploy_production(ctx):
    print("Deploying to production environment...")
    shmake.shell('echo "Production deployment started"', prefix='[PROD]')

# Parent deploy command
shmake.command(
    name = "deploy",
    help = "Deploy to different environments"
)

# Sub-commands for deployment
shmake.sub_command(
    path = ["deploy", "staging"],
    help = "Deploy to staging environment",
    action = deploy_staging
)

shmake.sub_command(
    path = ["deploy", "production"],
    help = "Deploy to production environment", 
    action = deploy_production
)

# ============================================================================
# FILE TIMESTAMP EXAMPLES
# ============================================================================
def demo_file_timestamps(ctx):
    """Demonstrates file timestamp functions newest_ts and oldest_ts"""
    print("=== File Timestamp Demo ===")
    
    # Create some test files for demonstration
    shmake.shell('echo "File 1" > demo1.txt && sleep 1')
    shmake.shell('echo "File 2" > demo2.txt && sleep 1') 
    shmake.shell('echo "File 3" > demo3.txt')
    
    # Example 1: Get newest timestamp from a single glob pattern
    newest = shmake.newest_ts('demo*.txt')
    print("Newest file timestamp:", newest)
    
    # Example 2: Get oldest timestamp from a single glob pattern  
    oldest = shmake.oldest_ts('demo*.txt')
    print("Oldest file timestamp:", oldest)
    
    # Create files in different directories for list example
    shmake.shell('mkdir -p src logs')
    shmake.shell('echo "Source code" > src/main.go && sleep 1')
    shmake.shell('echo "Log entry" > logs/app.log')
    
    # Example 3: Use list of globs to check multiple patterns
    newest_multi = shmake.newest_ts(['demo*.txt', 'src/*.go', 'logs/*.log'])
    oldest_multi = shmake.oldest_ts(['demo*.txt', 'src/*.go', 'logs/*.log'])
    
    print("Newest across all patterns:", newest_multi)
    print("Oldest across all patterns:", oldest_multi)
    
    # Show timestamp difference in seconds
    diff_seconds = newest - oldest
    print("Time difference:", diff_seconds, "seconds")
    
    # Clean up demo files
    shmake.shell('rm -f demo*.txt && rm -rf src logs')
    print("Demo files cleaned up")

shmake.command(
    name = "timestamps",
    help = "Demonstrate file timestamp functions newest_ts and oldest_ts",
    action = demo_file_timestamps
)

def demo_build_cache(ctx):
    """Advanced example using file timestamps for build caching"""
    print("=== Build Cache with File Timestamps ===")
    
    # Create source files
    shmake.shell('mkdir -p src')
    shmake.shell('echo "package main" > src/main.go')
    shmake.shell('echo "// helper functions" > src/utils.go')
    
    # Simulate build output
    shmake.shell('mkdir -p bin')
    shmake.shell('echo "fake binary" > bin/app')
    
    # Get timestamps
    source_newest = shmake.newest_ts('src/*.go')
    
    # Check if binary exists and get its timestamp
    check_result = shmake.shell('test -f bin/app && echo "exists"')
    binary_exists = check_result.stdout.strip() == "exists"
    
    if binary_exists:
        binary_ts = shmake.oldest_ts('bin/app')
    else:
        binary_ts = 0
    
    print("Source files newest timestamp:", source_newest)
    print("Binary timestamp:", binary_ts if binary_exists else "N/A (binary doesn't exist)")
    
    # Determine if rebuild is needed
    needs_rebuild = not binary_exists or source_newest > binary_ts
    
    if needs_rebuild:
        print("Rebuild needed - sources are newer than binary")
        shmake.shell('echo "Rebuilding..." && sleep 1 && echo "fake binary $(date)" > bin/app', prefix='[BUILD]')
        print("Build completed!")
    else:
        print("Build up-to-date - no rebuild necessary")
    
    # Clean up
    shmake.shell('rm -rf src bin')
    print("Demo files cleaned up")

shmake.command(
    name = "build-cache",
    help = "Advanced example using timestamps for intelligent build caching",
    action = demo_build_cache
)

# ============================================================================
# HELP COMMAND
# ============================================================================
def show_examples(ctx):
    """Show available example commands"""
    print("""
=== Available shmake Feature Demos ===

Basic Commands:
  ./shmake shell       - Shell command execution
  ./shmake async       - Async operations with start/wait  
  ./shmake pool        - Pool-based concurrent operations
  ./shmake templates   - String templating features
  ./shmake versioning  - Version storage and caching with cache instances
  ./shmake diff        - Version comparison with cache.diff()
  ./shmake timestamps  - File timestamp functions newest_ts and oldest_ts
  ./shmake build-cache - Advanced build caching using file timestamps

Advanced Examples:
  ./shmake build [target] --verbose --parallel
                       - Comprehensive build with args/flags
  ./shmake deploy staging     - Deploy to staging
  ./shmake deploy production  - Deploy to production

Try running any of these commands to see shmake features in action!
""")

shmake.command(
    name = "examples",
    help = "Show all available feature demonstrations",
    action = show_examples
)
