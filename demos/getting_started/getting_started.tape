Output demos/getting_started/getting_started.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Hide
Type "demos/getting_started/reset.sh" Enter
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demos/getting_started/configs` Enter
Sleep 1s
Show

# list of modules
Screenshot demos/getting_started/modules.png
# sleep a bit because otherwise vhs seems to take a screenshot of the init in
# the next step....
Sleep 0.5s

# init module a
Type "i" Sleep 0.5s
Screenshot demos/getting_started/init.png

# go to module a's page
Type "m" Sleep 0.5s
Screenshot demos/getting_started/module.png

# create a plan
Type "p" Sleep 0.5s
Screenshot demos/getting_started/plan.png

# apply plan
Type "a" Sleep 0.5s
Screenshot demos/getting_started/confirm.png

Type "y" Sleep 0.5s
Screenshot demos/getting_started/apply.png

# wait for apply to finish and state to be reloaded
Sleep 3s

# go to workspace page and cycle through tabs to the resources tab
Type "w" Sleep 0.5s Tab Tab
Screenshot demos/getting_started/resources.png
