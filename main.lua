local shmake = require("shmake.main")
local files = require("shmake.files")

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

local cli = shmake.command('shmake', { usage = "make shmake"})

local flagcmd = cli:command("flags", { usage = "show flags and args" })
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
flagcmd:command("subber")
    :action(function(flags)
        print(dump(flags))
    end)

-- or with the global function sub_command giving the path to use
cli:sub_command({"flags", "subcommand"})
    :flag("other-flag")
    :action(function(flags)
        print(dump(flags))
    end)

-- subcommands can have subcommands
cli:sub_command({"flags", "subcommand", "subsub"})
    :flag("third-flag")
    :action(function(flags)
        print(dump(flags))
    end)

-- define a variable as global and it will automatically be accessible for template expansion
someDir = current_dir.."/foobar"
cli:command("templates")
    :action(function()
        print(shmake.string([[
        current dir is {{.current_dir}}
        some dir is {{.someDir}}
        other var is {{.other_var}}
        ]], { other_var = "something" }))
    end)

cli:command("async"):action(function()
    shmake.async(function() shmake.shell('sleep 1; echo "first"', { prefix = 'one' }) end)
    shmake.async(function() shmake.shell('sleep 2; echo "second"', { prefix = 'two' }) end)
    shmake.wait()
    shmake.shell('echo "third"')
end)

cli:command("watch"):action(function()
    shmake.watch('./file3', function()
        print('touched file3, deleting file2')
        files.delete('file2')
    end)
    shmake.watch('./file4', function()
        print('touched file4, deleting file1')
        files.delete('file1')
    end)
    shmake.wait()
end)

cli:command("run"):action(function()
    shmake.shell('echo "running"')
end)

cli:command("output"):action(function()
    local output = shmake.shell('echo "running"')
    print("output is "..output)
    -- or use globals to allow templates
    out = shmake.shell('echo "running again"')
    print(shmake.string('output is {{.out}}'))
end)

cli:command("start"):action(function()
    shmake.watch('./file', function()
        local pool = shmake.pool()
        print('start pinging')
        pool:run(function() shmake.shell('ping google.com', { prefix="google "}) end)
        pool:run(function() shmake.shell('ping telness.se', {prefix="telness"}) end)
        pool:wait()
    end)
end)

cli:command("pool"):action(function()
    local pool = shmake.pool()
    pool:run(function() shmake.shell('ping google.com') end)
    pool:run(function() shmake.shell('ping telness.se') end)
    pool:wait()
end)

cli:command('files'):action(function() end)

cli:sub_command({'files', 'create'})
    :arg('file_name')
    :action(function(flags, args)
        files.mkdir('.files')
        files.chdir('.files')
        files.write(args[1], 'content')
        files.copy({from=args[1], to='another_file'})
        files.popdir()
    end)

cli:sub_command({'files', 'delete'})
    :arg('file_name')
    :action(function(flags, args)
        files.chdir('.files')
        files.delete(args[1])
        files.popdir()
        files.delete('.files')
    end)

cli:command('mod'):action(function()
    shmake.with_version({ name = "go.mod", int_version = files.newest_ts("go.mod")}, function()
        shmake.shell([[ go mod tidy ]])
    end)
end)

cli:run()
