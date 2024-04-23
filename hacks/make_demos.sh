vhs demos/modules.tape
vhs demos/runs.tape
vhs demos/run.tape
vhs demos/state.tape

MODULES_GIF_URL=$(vhs publish demos/modules.gif)
sed -i -e "s|\(\!\[Modules demo\]\).*|\1($MODULES_GIF_URL)|" README.md

RUNS_GIF_URL=$(vhs publish demos/runs.gif)
sed -i -e "s|\(\!\[Runs demo\]\).*|\1($RUNS_GIF_URL)|" README.md

RUN_GIF_URL=$(vhs publish demos/run.gif)
sed -i -e "s|\(\!\[Run demo\]\).*|\1($RUN_GIF_URL)|" README.md

STATE_GIF_URL=$(vhs publish demos/state.gif)
sed -i -e "s|\(\!\[State demo\]\).*|\1($STATE_GIF_URL)|" README.md
