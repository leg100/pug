Output demos/modules/modules.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Hide
Type "demos/modules/reset.sh" Enter
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demos/modules/configs` Enter
Sleep 1s
Show

Sleep 1s

Ctrl+a Sleep 0.5s Type "i" Sleep 2s
Type "i" Sleep 2s
Type "M"
Ctrl+a Sleep 0.5s Type "f" Sleep 2s
Ctrl+a Sleep 0.5s Type "v" Sleep 2s
Type "T" Sleep 2s Enter Sleep 2s
