Output demos/run/run.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Hide
Type "demos/run/reset.sh" Enter
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demos/run/configs` Enter
Sleep 1s
# run terraform init to initialize module but don't show viewer this
Type "i" Sleep 2s
Type "M"
Show

Sleep 1s Type "p" Sleep 2s
Sleep 0.5s Type "a" Sleep 1s Type "y" Sleep 5s
