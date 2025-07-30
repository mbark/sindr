local shell = require("shmake.shell")
local files = require("shmake.files")
local cache = require("shmake.cache")
local git = require("shmake.git")
local yarn = require("shmake.yarn")
local run = require("shmake.run")

local cli = shmake.new()

cli:command("start", { usage = "start pinging" })
    :action(function()
        shell.start({
            foo = { cmd = [[ ping google.com ]], watch = "./file" },
            bar = { cmd = [[ ping telness.se ]], watch = "./file2" }
        })
    end)

-- cli:command("mod", { usage = "go mod tidy on git change" })
--     :action( function()
--         cache.with_version({ name = "go.mod", int_version = files.newest_ts("go.mod")}, function()
--             shell.run([[ go mod tidy ]])
--         end)
--     end)

cli:run()
