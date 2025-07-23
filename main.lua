local shmake = require("shmake.main")
local shell = require("shmake.shell")
local files = require("shmake.files")
local cache = require("shmake.cache")
local git = require("shmake.git")
local yarn = require("shmake.yarn")
local run = require("shmake.run")

shmake.var({ name = 'Project', value = './../../..' })
shmake.var({ name = 'BackendPath', value = '{{.Project}}/backend' })
shmake.var({ name = 'ToolsBin', value = '{{.BackendPath}}/backend' })
shmake.var({ name = 'Path', value = shell.output('echo $PATH') })

local dev = shmake.env { name = "dev", default = true }
local prod = shmake.env { name = "prod" }

prod_clean = shmake.cmd('prod_clean', function()
    files.delete('*.pyc')
end, { env = prod })

shmake.cmd('clean', function()
    prod_clean()
    files.delete('file')
    files.delete('file2')
    files.delete({ files = "some_dir", only_directories = true })
    files.delete({ files = "nested", only_directories = true })
end)

shmake.cmd('clean_watch', function()
    run.watch({
        file = { fn = files.delete, args = 'file2', watch = './file3' }
    })
end)

shmake.cmd('install', function()
    files.chdir('./examples/yarn')
    yarn.install()
    yarn.run('prettier -w package.json')
    files.popdir()
end)

shmake.cmd('async', function()
    run.async(shell.run, 'sleep 2; echo "{{.Project}}"; echo "first"')
    run.async(shell.run, 'sleep 2; echo "second"')
    run.await()
    shell.run('echo "fourth"')
end)

shmake.cmd('a_script', function()
    shell.run([[ touch "file" ]])
    files.copy({ from = 'file', to = 'file2' })
    files.mkdir({ dir = 'nested/directory', all = true })
end)

shmake.cmd('start', function()
    shell.start({
        foo = { cmd = [[ ping google.com ]], watch = "./file" },
        bar = { cmd = [[ ping telness.se ]], watch = "./file2" }
    })
end)

shmake.cmd('mod', function()
    gomod_modtime = tostring(files.newest_ts("go.mod"))
    if cache.diff({ name = "go.mod", version = gomod_modtime }) then
        shell.run([[ go mod tidy ]])
        cache.store({ name = "go.mod", version = gomod_modtime })
    end
end)

shmake.cmd('proto', function()
    if files.newest_ts('*.proto', '{{.BACKEND_PATH}}/bin/inject') > files.oldest_ts('*.pb.go') then
        shell.run([[
            go mod vendor
            {{.PROTOC}} \
                -I {{.PROTO_INC}} \
                --validate_out="lang=go:.." \
                --twirp_out=.. \
                --go_out=.. \
                {{.RPC_V1_SRC}}
            {{.PROTOC}} \
                -I {{.PROTO_INC}} \
                --validate_out="lang=go:.." \
                --twirp_out=.. \
                --go_out=.. \
                {{.RPC_V2_SRC}}
            {{.PROTOC}} \
                -I {{.PROTO_INC}} \
                --validate_out="lang=go:.." \
                --go_out=.. \
                --gotemplate_out=all=true:. \
                {{.PROTO_DIR}}/telness/event/v1/*.proto
            {{.TOOLS_BIN}}/inject {{.BACKEND_PATH}}/src/telness/internal/rpc/*.pb.go
        ]])
        files.delete('{{.PROTO_VND}}')
    end
end)
