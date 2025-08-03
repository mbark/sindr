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

local startcmd = cli:command("start", { usage = "start pinging" })
    :flag("some-value", { usage = "pass some flag" })
    :string_flag("other-value", { default = "foobar" })
    :int_flag("some-int", { default = 5 })
    :bool_flag("is-bool", { default = false })
    :action(function(flags)
        print(dump(flags))
    end)

-- define sub commands either by using command on the variable
startcmd:command("subber")
    :action(function(flags)
        print(dump(flags))
    end)

-- or with the global function sub_command giving the path to use
cli:sub_command({"start", "subcommand"})
    :flag("other-flag")
    :action(function(flags)
        print(dump(flags))
    end)

cli:sub_command({"start", "subcommand", "subsub"})
    :flag("third-flag")
    :action(function(flags)
        print(dump(flags))
    end)

cli:run()
