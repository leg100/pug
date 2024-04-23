Output demos/output/runs.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Source demos/tapes/setup.tape

Ctrl+a Sleep 0.5s Type "i" Sleep 3s
Ctrl+a Sleep 0.5s Type "p" Sleep 4s

Ctrl+a Sleep 0.5s Type "a" Sleep 1s
Screenshot demos/output/apply_awaiting_confirmation.png
Type "y" Sleep 2s
Screenshot demos/output/applied_runs.png

Type "R" Sleep 3s
