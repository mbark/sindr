local shell = require("shmake.shell")
local files = require("shmake.files")
local cache = require("shmake.cache")
local run = require("shmake.run")
local string = require("shmake.string")

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
    :arg("some arg")
    :int_arg("int arg")
    :action(function(flags, args)
        print(dump(flags))
        print(dump(args))
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

-- define a variable as global and it will automatically be accessible for template expansion
someDir = current_dir.."/foobar"
cli:command("templates")
    :action(function()
        print(string.template([[
        current dir is {{.current_dir}}
        some dir is {{.someDir}}
        other var is {{.other_var}}
        ]], { other_var = "something" }))
    end)

cli:command("async"):action(function()
    run.async(function() shell.run('sleep 1; echo "first"') end)
    run.async(function() shell.run('sleep 2; echo "second"') end)
    run.await()
    shell.run('echo "third"')
end)

cli:command("watch"):action(function()
    run.async(function() shell.run('sleep 1; echo "first"') end)
    run.async(function() shell.run('sleep 2; echo "second"') end)
    run.await()
    shell.run('echo "third"')
end)

cli:command("clean"):action(function()
    run.watch('./file3', function() files.delete('file2') end)
    run.watch('./file4', function() files.delete('file1') end)
    run.await()
end)

cli:run()
