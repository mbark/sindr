local shell = require("shmake.shell")
local files = require("shmake.files")
local cache = require("shmake.cache")
local git = require("shmake.git")
local yarn = require("shmake.yarn")
local run = require("shmake.run")

function dump(o)
   if type(o) == 'table' then
      local s = '{ '
      for k,v in pairs(o) do
         if type(k) ~= 'number' then k = '"'..k..'"' end
         s = s .. '['..k..'] = ' .. dump(v) .. ','
      end
      return s .. '} '
   else
      return tostring(o)
   end
end

local cli = shmake.new('shmake', { usage = "make shmake"})

cli:command("start", { usage = "start pinging" })
    :flag("some-value", { usage = "pass some flag", required = true })
    :string_flag("other-value", { default = "foobar" })
    :int_flag("some-int", { default = 5 })
    :bool_flag("is-bool", { default = false })
    :action(function(flags)
        print(dump(flags))
    end)

cli:command("start something", { usage = "start pinging" })
    :action(function(flags)
        print("something running")
    end)

-- cli:command("mod", { usage = "go mod tidy on git change" })
--     :action( function()
--         cache.with_version({ name = "go.mod", int_version = files.newest_ts("go.mod")}, function()
--             shell.run([[ go mod tidy ]])
--         end)
--     end)

cli:run()
