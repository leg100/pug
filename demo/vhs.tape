Output demo/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Hide
Type "demo/reset.sh" Enter
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demo` Enter
Sleep 1s
Show

# show unintialized modules
Sleep 1s

# init all modules
Ctrl+a Sleep 0.5s Type "i"
# we're taken to the init task group page
Sleep 0.5s
# preview output for several tasks
Down Sleep 0.5s Down Sleep 0.5s Down Sleep 0.5s

# go back to modules
Type "m"
# validate all modules
Type "v"
# we're taken to the validate task group page
Sleep 0.5s
# enable task info side panel
Type "I" Sleep 0.5s
# take screen shot of tasks
Screenshot demo/tasks.png
# preview output for several tasks
Down Sleep 0.5s Down Sleep 0.5s Down Sleep 0.5s
# disable task info side panel
Type "I" Sleep 0.5s

# go back to modules
Type "m" Sleep 0.5s
# show help
Type "?" Sleep 0.5s
# run plan on all modules
Type "p"
# we're taken to the plan task group page
Sleep 1s
# hide help
Type "?"
# preview output for several tasks
Down Sleep 0.5s Down Sleep 0.5s Down Sleep 0.5s

# apply all plans
Ctrl+a Sleep 0.5s Type "a"
# accept confirmation
Sleep 0.5s Type "y"
# preview output for several tasks
Sleep 2s Down Sleep 1s Down Sleep 1s Down Sleep 1s
# take screen shot of task groups
Screenshot demo/task_groups.png

# go to workspaces
Type "w" Sleep 1s
# take screen shot of workspaces
Screenshot demo/workspaces.png
# filter dev workspaces
Type "/" Sleep 0.5s Type "dev" Sleep 0.5s Enter
# select all dev workspaces and auto-apply
Ctrl+a Sleep 0.5s Type "a"
# accept confirmation
Sleep 0.5s Type "y"
# preview output for several tasks
Sleep 2s Down Sleep 1s Down Sleep 1s

# go back to modules
Type "m" Sleep 0.5s
# take screen shot of modules (sleep to ensure page doesn't switch too soon)
Screenshot demo/modules.png Sleep 0.5s

# go to state for current module
Type "s" Sleep 0.5s
# see preview for several resources
Sleep 0.5s Down Sleep 0.5s Down Sleep 0.5s Down Sleep 0.5s

# move current resource
Type "M" Sleep 0.5s Backspace 6 Sleep 0.5s Type "giraffe[99]" Sleep 0.5s Enter
# look at task output for a bit
Sleep 1s

# go back to state
Escape Sleep 0.5s
# take screen shot of state
Screenshot demo/state.png

# delete another resource
Down Sleep 0.5s
Type "D"
# confirm deletion
Sleep 0.5s Type "y" Sleep 0.5s
# look at task output for a bit
Sleep 1s

# go back to state
Escape Sleep 0.5s

# select all resources
Ctrl+a Sleep 0.5s
# taint all resources
Ctrl+t
# we're taken to the taint task group page
Sleep 0.5s
# preview output for several tasks
Down Sleep 0.5s Down Sleep 0.5s Down Sleep 0.5s

# go back to state
Escape Sleep 0.5s

# select all resources
Ctrl+a Sleep 0.5s
# taint all resources
Ctrl+u
# we're taken to the untaint task group page
Sleep 0.5s
# preview output for several tasks
Down Sleep 0.5s Down Sleep 0.5s Down Sleep 0.5s

# go back to state
Escape Sleep 0.5s

# select several resources
Space Down Sleep 0.5s Space Down Sleep 0.5s Down Space Sleep 0.5s
# target resources in a plan
Type "p"
# look at plan task page for a bit
Sleep 1s
# apply plan
Type "a"
# accept confirmation
Sleep 0.5s Type "y"
# look at apply task page for a bit
Sleep 1.5s
