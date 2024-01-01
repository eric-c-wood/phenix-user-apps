## win-cmds

Aggregates a collection of windows commands into one batch script to ran when the virtual machine boots.  Each set of commands is described using Python's mako library.  Any reboot commands are added at the end of the batch. 

### configuration

To apply a specific template to a virtual machine, add a `win-<version>-cmds` key to a `labels` key in the topology.  The values for the `win-<version>-cmds` key will be a comma separated list of mako template names without the `mako` extension.  The `<version>` will represent the windows version (e.g. xp,7,8,10, etc).  Look at the `example-configs` folder for an understanding of how to configure this user app.





