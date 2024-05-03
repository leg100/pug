Output demos/runs/runs.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Framerate 24
Set Padding 5

Type "# Welcome to the demo...." Sleep 0.5s

Hide
Type "demos/runs/reset.sh" Enter
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demos/runs/configs` Enter
Sleep 1s
# run terraform init to initialize modules but don't show viewer this
Ctrl+a Type "i" Ctrl+a Sleep 3s
Show

Sleep 2s

# select all modules and create plan for each
Ctrl+a Sleep 0.5s Type "p" Sleep 4s

# create an apply for all, and enter 'y' to confirm
Sleep 0.5s Type "a" Sleep 1s
Screenshot demos/runs/apply_awaiting_confirmation.png
Type "y" Sleep 2s

# navigate to global runs listing
Type "R" Sleep 3s
Screenshot demos/runs/applied_runs.png

Ctrl+C Sleep 0.5s Type "y"
Type "That's the end of the demo...." Sleep 0.5s
