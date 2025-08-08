def build(ctx):
    print("building", ctx.args.target)

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
    name = "deploy",
    help = "Deploy to an environment",
    action = deploy,
    args = ["env"]
)
