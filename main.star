load('cmd/cmd.star', 'cmd')

shmake.load_package_json('package.json')

def build(ctx):
    out = shmake.shell('echo "foobar"', prefix='hej')
    print('output', out)
    print("building", ctx.args.target, 'snake_cased', ctx.flags.some_flag, 'by flag-name', ctx.flags['some-flag'])

shmake.cli(
    name = "shmake",
    usage = "A sample CLI tool"
)

# Register commands
shmake.command(
    name = "build",
    help = "Build the project",
    action = build,
    args = ["target"],
    flags = {
        "some-flag": {
            "type": "bool",
            "default": False,
        }
    }
)

shmake.command(
    name = "build_cmd",
    help = "imported from cmd.star",
    action = cmd,
    category = "cmd",
)

shmake.command(
    name = "flags",
    action = lambda ctx:
        print("flags")
)

shmake.sub_command(
    path = ["flags", "subber"],
    action = lambda ctx:
        print('running sub command')
)

def start(ctx):
    shmake.start(lambda: shmake.shell('sleep 1; echo "first"', prefix='one'))
    shmake.start(lambda: shmake.shell('sleep 2; echo "second"', prefix='two'))
    shmake.wait()
    shmake.shell('echo "third"')

shmake.command(
    name='async',
    action=start,
)

def pooled_start():
    pool = shmake.pool()
    print('start pinging')
    pool.run(lambda: shmake.shell('ping google.com', prefix='google '))
    pool.run(lambda: shmake.shell('ping telness.se', prefix='telness'))
    pool.wait()

print(current_dir)
# defining a global variable makes it available for string templating
some_dir = "foobar"
shmake.command(
    name = "string",
    help = "Deploy to an environment",
    action = lambda ctx:
        print(shmake.string('''
            global variable: {{.some_dir}}
            dict variable: {{.other_var}}
            ''', other_var='other variable')),
    args = ["env"]
)
