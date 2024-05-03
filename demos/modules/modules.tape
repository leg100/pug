Output demos/modules/modules.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Type "# Welcome to the demo...." Sleep 0.5s

Hide
Type "demos/modules/reset.sh" Enter
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demos/modules/configs` Enter
Sleep 1s
Screenshot demos/modules/modules_listing.png
Show

Sleep 1s

# init all modules
Ctrl+a Sleep 0.5s Type "i" Ctrl+a Sleep 2s
# init first module, again, so that we are taken to its task page
Type "i" Sleep 2s
Type "M"
# select all
Ctrl+a
# format all
Sleep 0.5s Type "f" Sleep 2s
# validate all
Sleep 0.5s Type "v" Sleep 2s
Type "T" Sleep 2s Enter Sleep 2s

Ctrl+C Sleep 0.5s Type "y"
Type "That's the end of the demo...." Sleep 0.5s
