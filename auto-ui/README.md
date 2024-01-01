## auto-ui

The auto-ui app sends a linear list of commands/actions to one or more virtual machines in a Phenix experiment.  The list of commands/actions is contained in a YAML configuration file.  The YAML configuration file can be automatically generated based on a Phenix topology as part of the phenix-app-autoui user app.   

### command list 

**gui-terminal-commands**: Opens a terminal in a GUI and sends a list of commands to the terminal.  The `desktop-image` parameter references an image used to identify when a `desktop` is visible.  The `terminal-window-image` parameter references an image used to identify when a command line terminal is visible.  After all commands have been ran, the terminal is closed.
**linux-terminal-commands**: Sends a list of commands to a non-GUI linux terminal.  It is assumed that a user is already logged in.  There are other commands to assist with logins.
**login-linux**: Logs into a non-GUI linux terminal. The `linux-login-prompt` parameter references an image used to identify when a linux login prompt is visible.  The `linux-after-login-prompt` parameter references an image used to identify when a user has successfully logged in.  
**send-special-keys**: Sends keyboard input that is a combination of more than one character such as `tab` and `enter`.
**sleep**: Pauses execution of the list of commands/actions for a user defined period of time. (i.e the `timeout` parameter)  
**wait-until**: Pauses execution of the list of commands/actions until a reference image (i.e. the `reference-image` parameter)

#### coming soon

**send-kb-shortcut**: Used to emulate combination key presses (i.e. `CTRL+ALT+DELETE`, `ALT_L+ENTER`)
**click-image**: Clicks the center of a reference image  
**right-click-image**: Right clicks the center of a reference image
**double-click-image**: Double clicks the center of a reference image  
**move-to-image**: Moves the mouse cursor to the center of a reference image

### running from the command line

The current implementation only includes two parameters namely the name of a running experiment `-exp` and a YAML configuration file `-config` with a linear list of actions/commands to execute.

`auto-ui -exp <experiment_name> -config <YAML configuration file>`



