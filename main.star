def build(ctx):
    out = shmake.shell('echo "foobar"', prefix='hej')
    print('output', out)
    print("building", ctx.args.target, 'snake_cased', ctx.flags.some_flag, 'by flag-name', ctx.flags['some-flag'])

def deploy(ctx):
    print("deploying", ctx.args.env)

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
    name = "flags",
    action = lambda ctx:
        print("flags")
)

shmake.sub_command(
    path = ["flags", "subber"],
    action = lambda ctx:
        print('running sub command')
)

# defining a global variable makes it available for string templating
someDir = "foobar"
shmake.command(
    name = "string",
    help = "Deploy to an environment",
    action = lambda ctx:
        print(shmake.string('''
            global variable: {{.someDir}}
            dict variable: {{.other_var}}
            ''', other_var='other variable')),
    args = ["env"]
)
