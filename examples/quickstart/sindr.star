cli(
    name = "cli_name",
    usage = "some usage text"
)

def a_command(ctx):
    res = shell(string('echo "{{.text}}"',text=ctx.flags.text))
    print(res.stdout)

command(
    name = "a_command",
    action = a_command,
    flags = {
        "text": {
            "type": "string",
            "default": "hello from sindr",
            "help": "text to echo"
        },
    },
)
