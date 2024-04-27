Output demos/workspaces/workspaces.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Hide
Type "demos/workspaces/reset.sh" Enter
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demos/workspaces/configs` Enter
Sleep 1s
# run terraform init to initialize modules but don't show viewer this
Ctrl+a Type "i" Sleep 3s
Show

Sleep 1s

# switch to global workspace listing
Type "W" Sleep 3s
# trigger a plan on all workspaces
Ctrl+a Sleep 0.5s Type "p" Sleep 2s
# trigger an an apply on each workspace, and enter 'y' to confirm
Ctrl+a Sleep 0.5s Type "a" Sleep 1s
Type "y" Sleep 4s
# switch current workspace on module a from default to dev
Down Sleep 1s Type "C" Sleep 2s
# switch to global module listing, showing the current workspace has changed for module a
Type "M" Sleep 2s
