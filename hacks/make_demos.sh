vhs demos/tapes/modules.tape
vhs demos/tapes/runs.tape
vhs demos/tapes/run.tape
vhs demos/tapes/state.tape

MODULES_GIF_URL=$(vhs publish demos/output/modules.gif)
sed -i -e "s|\(\!\[Modules demo\]\).*|\1($MODULES_GIF_URL)|" README.md

RUNS_GIF_URL=$(vhs publish demos/output/runs.gif)
sed -i -e "s|\(\!\[Runs demo\]\).*|\1($RUNS_GIF_URL)|" README.md

RUN_GIF_URL=$(vhs publish demos/output/run.gif)
sed -i -e "s|\(\!\[Run demo\]\).*|\1($RUN_GIF_URL)|" README.md

STATE_GIF_URL=$(vhs publish demos/output/state.gif)
sed -i -e "s|\(\!\[State demo\]\).*|\1($STATE_GIF_URL)|" README.md
