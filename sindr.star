# Configure CLI metadata
cli(
    name = "sindr",
    usage = "Comprehensive demo of sindr features and functions"
)


def demo_shell(ctx):
    """Demonstrates shell() with different options"""
    print("=== Shell Command Demo ===")
    
    # Basic shell command
    result = shell('echo "Hello from sindr!"')
    print("Shell output:", result.stdout)
    print("Command success:", result.success)
    print("Exit code:", result.exit_code)
    
    # Shell command with prefix
    shell('echo "This has a prefix"', prefix='[DEMO]')
    
    # Multiple commands
    shell('echo "Command 1" && echo "Command 2"')

command(
    name = "shell",
    usage = "Demonstrate shell command execution",
    action = demo_shell
)

load_package_json('package.json')

# ============================================================================
# ASYNC OPERATIONS EXAMPLE
# ============================================================================
def demo_async(ctx):
    """Demonstrates async operations with start() and wait()"""
    print("=== Async Operations Demo ===")
    
    # Start multiple operations concurrently
    start(lambda: shell('sleep 1 && echo "Task 1 completed"', prefix='[TASK1]'))
    start(lambda: shell('sleep 2 && echo "Task 2 completed"', prefix='[TASK2]'))
    start(lambda: shell('sleep 1.5 && echo "Task 3 completed"', prefix='[TASK3]'))
    
    print("Started 3 async tasks, waiting for completion...")
    wait()
    print("All async tasks completed!")

command(
    name = "async",
    usage = "Demonstrate async operations with start/wait",
    action = demo_async
)

# ============================================================================
# POOL OPERATIONS EXAMPLE
# ============================================================================
def demo_pool(ctx):
    """Demonstrates pool-based concurrent operations"""
    print("=== Pool Operations Demo ===")
    
    p = pool()
    
    # Add tasks to the pool
    p.run(lambda: shell('echo "Pool task 1" && sleep 1', prefix='[POOL1]'))
    p.run(lambda: shell('echo "Pool task 2" && sleep 1', prefix='[POOL2]'))
    p.run(lambda: shell('echo "Pool task 3" && sleep 1', prefix='[POOL3]'))
    
    print("Pool tasks started, waiting...")
    p.wait()
    print("Pool tasks completed!")

command(
    name = "pool",
    usage = "Demonstrate pool-based concurrent operations",
    action = demo_pool
)

# ============================================================================
# STRING TEMPLATING EXAMPLE
# ============================================================================
def demo_templates(ctx):
    """Demonstrates string templating with global and local variables"""
    print("=== String Templating Demo ===")
    
    # Using global variables in template
    template1 = string('''
Project: {{.project_name}}
Version: {{.version}}
Build Directory: {{.build_dir}}
''', project_name='sindr', version='0.0.1', build_dir=current_dir)
    print("Template with global variables:")
    print(template1)
    
    # Using local variables in template
    template2 = string('''
Environment: {{.env}}
Database: {{.db_host}}:{{.db_port}}
Debug Mode: {{.debug}}
''', env='production', db_host='localhost', db_port=5432, debug=False)
    
    print("Template with local variables:")
    print(template2)
    
    # Complex template with conditionals and loops
    template3 = string('''
{{if .enable_feature}}Feature is enabled!{{else}}Feature is disabled{{end}}
Services: {{range .services}}
  - {{.}}{{end}}
''', enable_feature=True, services=['web', 'api', 'worker'])
    
    print("Complex template:")
    print(template3)

command(
    name = "templates",
    usage = "Demonstrate string templating features",
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
        shell('sleep 2 && echo "Build completed!"')
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

command(
    name = "versioning",
    usage = "Demonstrate versioning and caching with cache instances",
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

command(
    name = "diff",
    usage = "Demonstrate version comparison with cache.diff()",
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
    timestamp_result = shell('date +%s')
    build_version = "build-" + timestamp_result.stdout.strip()
    c.set_version(name="current_build", version=build_version)
    
    # Create build script using templates
    build_script = string('''#!/bin/bash
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
        pool = pool()
        pool.run(lambda: shell('echo "Compiling frontend..." && sleep 1', prefix='[FRONTEND]'))
        pool.run(lambda: shell('echo "Compiling backend..." && sleep 2', prefix='[BACKEND]'))
        pool.run(lambda: shell('echo "Running tests..." && sleep 1.5', prefix='[TESTS]'))
        pool.wait()
    else:
        print("Running sequential build...")
        shell('echo "Compiling frontend..."', prefix='[BUILD]')
        shell('echo "Compiling backend..."', prefix='[BUILD]')
        shell('echo "Running tests..."', prefix='[BUILD]')
    
    # Final status
    shell(string('echo "Build {{.build_id}} completed successfully!"', build_id=build_version))

command(
    name = "build",
    usage = "Comprehensive build example with args and flags",
    action = comprehensive_build,
    args = ["target"],
    flags = [
        {
            "name": "verbose",
            "type": "bool",
            "default": False,
            "usage": "Enable verbose output"
        },
        {
            "name": "parallel",
            "type": "bool",
            "default": False,
            "usage": "Enable parallel build"
        }]
)

def string_slice_flag(ctx):
    print('strings:', ctx.flags.strings)
    print('ints:', ctx.flags.ints)

command(
    name = 'slice_flag',
    action=string_slice_flag,
    flags = [
        {
            "name": "strings",
            "type": "strings",
            "default": ["1","2"],
        },
        {
            "name": "ints",
            "type": "ints",
            "default": [1,2,3],
        },
        "simple",
 ]
)

def shell_string(ctx):
    res = shell('echo {{.string_flag}} {{.argument}}')
    print(res.stdout)
    print(ctx.string_flag)
    print(ctx.argument)

command(
    name = 'shell_string',
    action=shell_string,
    flags=[{'name': 'string_flag', 'type': 'bool'}],
    args=['argument'],
)

def exec(ctx):
    res = exec('python3', '''
print('Hello from python!')
''')
    print('output:', res.stdout)

command(
    name='exec',
    action=exec,
)

def dotenv_fn(ctx):
    dotenv()
    res = shell('echo $FOO')
    print('FOO is', res.stdout)
    res = shell('echo $EDITOR')
    print('EDITOR is', res.stdout)

    dotenv(overload=True)
    res = shell('echo $EDITOR')
    print('EDITOR is', res.stdout)

command(
    name='dotenv',
    action=dotenv_fn,
)

# ============================================================================
# SUB-COMMAND EXAMPLES
# ============================================================================
def deploy_staging(ctx):
    print("Deploying to staging environment...")
    shell('echo "Staging deployment started"', prefix='[STAGING]')

def deploy_production(ctx):
    print("Deploying to production environment...")
    shell('echo "Production deployment started"', prefix='[PROD]')

# Parent deploy command
command(
    name = "deploy",
    usage = "Deploy to different environments"
)

# Sub-commands for deployment
sub_command(
    path = ["deploy", "staging"],
    usage = "Deploy to staging environment",
    action = deploy_staging
)

sub_command(
    path = ["deploy", "production"],
    usage = "Deploy to production environment",
    action = deploy_production
)

# ============================================================================
# FILE TIMESTAMP EXAMPLES
# ============================================================================
def demo_file_timestamps(ctx):
    """Demonstrates file timestamp functions newest_ts and oldest_ts"""
    print("=== File Timestamp Demo ===")
    
    # Create some test files for demonstration
    shell('echo "File 1" > demo1.txt && sleep 1')
    shell('echo "File 2" > demo2.txt && sleep 1')
    shell('echo "File 3" > demo3.txt')
    
    # Example 1: Get newest timestamp from a single glob pattern
    newest = newest_ts('demo*.txt')
    print("Newest file timestamp:", newest)
    
    # Example 2: Get oldest timestamp from a single glob pattern  
    oldest = oldest_ts('demo*.txt')
    print("Oldest file timestamp:", oldest)
    
    # Create files in different directories for list example
    shell('mkdir -p src logs')
    shell('echo "Source code" > src/main.go && sleep 1')
    shell('echo "Log entry" > logs/app.log')
    
    # Example 3: Use list of globs to check multiple patterns
    newest_multi = newest_ts(['demo*.txt', 'src/*.go', 'logs/*.log'])
    oldest_multi = oldest_ts(['demo*.txt', 'src/*.go', 'logs/*.log'])
    
    print("Newest across all patterns:", newest_multi)
    print("Oldest across all patterns:", oldest_multi)
    
    # Show timestamp difference in seconds
    diff_seconds = newest - oldest
    print("Time difference:", diff_seconds, "seconds")
    
    # Clean up demo files
    shell('rm -f demo*.txt && rm -rf src logs')
    print("Demo files cleaned up")

command(
    name = "timestamps",
    usage = "Demonstrate file timestamp functions newest_ts and oldest_ts",
    action = demo_file_timestamps
)

def demo_build_cache(ctx):
    """Advanced example using file timestamps for build caching"""
    print("=== Build Cache with File Timestamps ===")
    
    # Create source files
    shell('mkdir -p src')
    shell('echo "package main" > src/main.go')
    shell('echo "// helper functions" > src/utils.go')
    
    # Simulate build output
    shell('mkdir -p bin')
    shell('echo "fake binary" > bin/app')
    
    # Get timestamps
    source_newest = newest_ts('src/*.go')
    
    # Check if binary exists and get its timestamp
    check_result = shell('test -f bin/app && echo "exists"')
    binary_exists = check_result.stdout.strip() == "exists"
    
    if binary_exists:
        binary_ts = oldest_ts('bin/app')
    else:
        binary_ts = 0
    
    print("Source files newest timestamp:", source_newest)
    print("Binary timestamp:", binary_ts if binary_exists else "N/A (binary doesn't exist)")
    
    # Determine if rebuild is needed
    needs_rebuild = not binary_exists or source_newest > binary_ts
    
    if needs_rebuild:
        print("Rebuild needed - sources are newer than binary")
        shell('echo "Rebuilding..." && sleep 1 && echo "fake binary $(date)" > bin/app', prefix='[BUILD]')
        print("Build completed!")
    else:
        print("Build up-to-date - no rebuild necessary")
    
    # Clean up
    shell('rm -rf src bin')
    print("Demo files cleaned up")

command(
    name = "build-cache",
    usage = "Advanced example using timestamps for intelligent build caching",
    action = demo_build_cache
)

# ============================================================================
# GLOB FUNCTION EXAMPLE
# ============================================================================
def demo_glob(ctx):
    """Demonstrates the glob function for file pattern matching"""
    print("=== Glob Function Demo ===")
    
    # Create some test files for demonstration
    shell('mkdir -p src test docs')
    shell('echo "package main" > src/main.go')
    shell('echo "package utils" > src/utils.go')
    shell('echo "import unittest" > test/test_main.py')
    shell('echo "import pytest" > test/test_utils.py')
    shell('echo "# README" > docs/README.md')
    shell('echo "# API Guide" > docs/api.md')
    shell('echo "config data" > config.json')
    shell('echo "more config" > settings.yml')
    
    # Example 1: Single glob pattern
    print("\n1. Finding all Go files:")
    go_files = glob('src/*.go')
    for file in go_files:
        print("  -", file)
    
    # Example 2: Multiple patterns using a list
    print("\n2. Finding Python and Markdown files:")
    script_files = glob(['test/*.py', 'docs/*.md'])
    for file in script_files:
        print("  -", file)
    
    # Example 3: Complex patterns with multiple extensions
    print("\n3. Finding all configuration files:")
    config_files = glob(['*.json', '*.yml', '*.yaml'])
    for file in config_files:
        print("  -", file)
    
    # Example 4: Using glob results in other operations
    print("\n4. Processing Go files individually:")
    for go_file in glob('src/*.go'):
        # Get file size instead of line count for simpler demo
        size_result = shell('stat -c %s ' + go_file + ' 2>/dev/null || stat -f %z ' + go_file)
        size = size_result.stdout.strip()
        print("  ", go_file, "is", size, "bytes")
    
    # Example 5: Combining glob with newest_ts/oldest_ts
    print("\n5. File age analysis:")
    all_source_files = glob(['src/*.go', 'test/*.py'])
    if all_source_files:
        newest_ts = newest_ts(['src/*.go', 'test/*.py'])
        oldest_ts = oldest_ts(['src/*.go', 'test/*.py'])
        print("  Found", len(all_source_files), "source files")
        print("  Newest file timestamp:", newest_ts)
        print("  Oldest file timestamp:", oldest_ts)
        print("  Age difference:", newest_ts - oldest_ts, "seconds")
    
    # Example 6: Conditional processing based on file existence
    print("\n6. Conditional operations:")
    makefile_candidates = glob(['Makefile', 'makefile', '*.mk'])
    if makefile_candidates:
        print("  Found build files:", makefile_candidates)
    else:
        print("  No build files found")
    
    # Example 7: File counting and statistics
    print("\n7. File statistics:")
    all_files = glob(['*', 'src/*', 'test/*', 'docs/*'])
    extensions = {}
    for file in all_files:
        if '.' in file:
            ext = file.split('.')[-1]
            extensions[ext] = extensions.get(ext, 0) + 1
    
    print("  Total files found:", len(all_files))
    print("  Files by extension:")
    for ext, count in extensions.items():
        print("    ." + ext + ":", count)
    
    # Clean up demo files
    shell('rm -rf src test docs config.json settings.yml')
    print("\nDemo files cleaned up")

command(
    name = "glob",
    usage = "Demonstrate glob function for file pattern matching",
    action = demo_glob
)

# ============================================================================
# HELP COMMAND
# ============================================================================
def show_examples(ctx):
    """Show available example commands"""
    print("""
=== Available sindr Feature Demos ===

Basic Commands:
  ./sindr shell       - Shell command execution
  ./sindr async       - Async operations with start/wait
  ./sindr pool        - Pool-based concurrent operations
  ./sindr templates   - String templating features
  ./sindr versioning  - Version storage and caching with cache instances
  ./sindr diff        - Version comparison with cache.diff()
  ./sindr timestamps  - File timestamp functions newest_ts and oldest_ts
  ./sindr glob        - File pattern matching with glob function
  ./sindr build-cache - Advanced build caching using file timestamps

Advanced Examples:
  ./sindr build [target] --verbose --parallel
                       - Comprehensive build with args/flags
  ./sindr deploy staging     - Deploy to staging
  ./sindr deploy production  - Deploy to production

Try running any of these commands to see sindr features in action!
""")

command(
    name = "examples",
    usage = "Show all available feature demonstrations",
    action = show_examples
)
