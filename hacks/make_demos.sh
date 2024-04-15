#vhs demos/modules.tape
#MODULES_GIF_URL=$(vhs publish demos/modules.gif)
#sed -i -e "s|\(\!\[Modules demo\]\).*|\1($MODULES_GIF_URL)|" README.md

vhs demos/runs.tape
RUNS_GIF_URL=$(vhs publish demos/runs.gif)
sed -i -e "s|\(\!\[Runs demo\]\).*|\1($RUNS_GIF_URL)|" README.md
