def cmd(ctx):
    print('cmd running')
    out = shell('echo "foobar"', prefix='hej')
    print('output', out)
