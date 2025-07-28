local shell = require("shmake.shell")
local files = require("shmake.files")
local cache = require("shmake.cache")
local git = require("shmake.git")
local yarn = require("shmake.yarn")
local run = require("shmake.run")

local cli = shmake.new()

cli:command("start", function()
    shell.start({
        foo = { cmd = [[ ping google.com ]], watch = "./file" },
        bar = { cmd = [[ ping telness.se ]], watch = "./file2" }
    })
end, { usage = "start pinging" })

cli:command("mod", function()
    cache.with_version({ name = "go.mod", int_version = files.newest_ts("go.mod")}, function()
        shell.run([[ go mod tidy ]])
    end)
end, { usage = "go mod tidy on git change" })


cli:run()
