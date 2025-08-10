def cmd(ctx):
    print('cmd running')
    out = shmake.shell('echo "foobar"', prefix='hej')
    print('output', out)
