Output demo/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Hide
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demo/dont_cost_money` Enter
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
# take screen shot of task group page
Screenshot demo/task_group.png
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

# go to workspaces
Type "w" Sleep 1s
# take screen shot of workspaces
Screenshot demo/workspaces.png
# filter dev workspaces
Type "/" Sleep 0.5s Type "dev" Sleep 0.5s Enter
# take screen shot of filter mode
Screenshot demo/filter.png Sleep 0.5s
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
Type "U"
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

# go to task listing
Type "t"
# enable task info sidebar
Type "I"
# take screenshot of task listing along with sidebar
Screenshot demo/tasks.png
Sleep 2s

# go to task groups listing
Type "T"
# take screenshot of task groups
Screenshot demo/task_groups.png
Sleep 2s

# go to logs
Type "l"
# take screenshot of logs
Screenshot demo/logs.png
Sleep 2s

# quit app and restart, this time with workspaces that cost money, to demonstrate infracost integration
Hide
Ctrl+c Type "y"
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demo/do_cost_money` Enter
Sleep 1s
Show

# show unintialized modules
Sleep 1s

# init all modules
Ctrl+a Sleep 0.5s Type "i"
# we're taken to the init task group page, wait for a few seconds for tasks to finish
Sleep 2s

# go to workspaces
Type "w" Sleep 2s
# select all workspaces
Ctrl+a Sleep 0.5s
# calculate cost
Type "$"
# user is taken to infracost task page, watch output for a few seconds
Sleep 6s
# take screen shot of infracost output (sleep to ensure page doesn't switch too soon)
Screenshot demo/infracost_output.png Sleep 0.5s

# go back to workspaces
Type "w" Sleep 0.5s
# take screen shot of workspaces (sleep to ensure page doesn't switch too soon)
Screenshot demo/workspaces_with_cost.png Sleep 0.5s
Sleep 3s

# quit app and restart, this time with modules with .tool-versions files, to demonstrate support for executing arbitrary programs, in this case, `asdf`
Hide
Ctrl+c Type "y"
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demo/multiple_versions` Enter
Sleep 1s
Show

# show unintialized modules
Sleep 1s

# select all modules
Ctrl+a Sleep 0.5s
# install terraform for each module
Type "x" Sleep 1s
Type "asdf install terraform" Sleep 1s
# take screen shot (sleep to ensure page doesn't switch too soon)
Screenshot demo/execute_asdf_install_terraform.png Sleep 0.5s
# Execute asdf
Enter
# Show user the output of asdf downloading terraform. This takes a long time,
# so sleep a bit, then hide and wait, then show and sleep a bit more.
Sleep 2s
Hide
Sleep 8s
Show
Sleep 4s
Screenshot demo/asdf_install_terraform_task_group.png Sleep 0.5s
