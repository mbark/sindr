def build(ctx):
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

shmake.command(
    name = "deploy",
    help = "Deploy to an environment",
    action = deploy,
    args = ["env"]
)
