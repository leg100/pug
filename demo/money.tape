Output demo/money.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Hide
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demo/workspaces_that_cost_money` Enter
Sleep 1s
Show

# show unintialized modules
Sleep 1s

# init all modules
Ctrl+a Sleep 0.5s Type "i"
# we're taken to the init task group page, wait for a few seconds for tasks to finish (it takes longer because it is copying hefty aws provider about
Sleep 3s

# go to workspaces
Type "w" Sleep 2s
# select all workspaces
Ctrl+a Sleep 0.5s
# calculate cost
Type "$"
# user is taken to infracost task page, watch output for a few seconds
Sleep 5s
# take screen shot of infracost output (sleep to ensure page doesn't switch too soon)
Screenshot demo/infracost_output.png Sleep 0.5s

# go back to workspaces
Type "w" Sleep 0.5s
# take screen shot of workspaces (sleep to ensure page doesn't switch too soon)
Screenshot demo/workspaces_with_cost.png Sleep 0.5s
Sleep 2s
