Output demos/state/state.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Type "# Welcome to the demo...." Sleep 0.5s

Hide
Type "demos/state/reset.sh" Enter
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demos/state/configs` Enter
Sleep 1s
# run terraform init to initialize module but don't show viewer this
Type "i" Sleep 2s
Type "M"
# run terraform plan and apply to create resources but don't show viewer this
Type "p" Sleep 2s
Type "a" Type "y" Sleep 2s
Type "W" Sleep 0.5s
Show

Sleep 1s Enter Sleep 0.5s Tab Sleep 0.5s Tab Sleep 1s

# move first resource
Alt+M Sleep 2s Backspace 6 Sleep 0.5s Type "giraffe[99]" Sleep 0.5s Enter Sleep 2s

# select first three resources
Type "s" Sleep 0.5s
Down Type "s" Sleep 0.5s
Down Type "s" Sleep 0.5s

# taint selected resources
Type "t" Sleep 2s

# untaint selected resources
Type "u" Sleep 2s

# delete selected resources
Type "D" Sleep 1s Type "y" Sleep 2s

# select three more resources
Type "s"
Up Type "s" Sleep 0.5s
Up Type "s" Sleep 0.5s

# create plan targeting selected resources
Type "p" Sleep 3s

Ctrl+C Sleep 1s Type "y"
Type "That's the end of the demo...." Sleep 0.5s
