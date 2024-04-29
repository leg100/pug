Output demos/tasks/tasks.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Hide
Type "demos/tasks/reset.sh" Enter
Type `TF_CLI_CONFIG_FILE=$PWD/mirror/mirror.tfrc go run main.go -w demos/tasks/configs` Enter
Sleep 1s
# run terraform init to initialize module but don't show viewer this
Type "i" Sleep 2s
Type "M"
# run terraform plan and apply to create resources but don't show viewer this
Type "p" Sleep 2s
Type "a" Type "y"
# show global task listing to user as the first thing they see
Type "T" Sleep 0.5s
Show

# wait for apply to complete (it runs for at least 3 seconds) and for state to be reloaded
Sleep 5s
# show task page for latest task, which should be the reloading of state, and scroll up
Enter Sleep 1s PageUp Sleep 0.5s PageUp 2 Sleep 0.5s PageUp Sleep 0.5s PageUp 2 Sleep 1s
# go back and show task page for penultimate task, which should be the apply
Type "T" Sleep 0.5s Down Sleep 0.5s Enter Sleep 1s
# scroll up apply output
Enter Sleep 1s PageUp Sleep 1s
Type "T" Sleep 1s
