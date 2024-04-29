vhs demos/modules/modules.tape
vhs demos/workspaces/workspaces.tape
vhs demos/runs/runs.tape
vhs demos/run/run.tape
vhs demos/state/state.tape
vhs demos/tasks/tasks.tape

# Publish animated gifs, sleeping in between publishing to be nice to charm's
# servers (seem to get an EOF when publishing the last tasks gif?), and update
# readme with links

MODULES_GIF_URL=$(vhs publish demos/modules/modules.gif)
sed -i -e "s|\(\!\[Modules demo\]\).*|\1($MODULES_GIF_URL)|" README.md

sleep 1

WORKSPACES_GIF_URL=$(vhs publish demos/workspaces/workspaces.gif)
sed -i -e "s|\(\!\[Workspaces demo\]\).*|\1($WORKSPACES_GIF_URL)|" README.md

sleep 1

RUNS_GIF_URL=$(vhs publish demos/runs/runs.gif)
sed -i -e "s|\(\!\[Runs demo\]\).*|\1($RUNS_GIF_URL)|" README.md

sleep 1

RUN_GIF_URL=$(vhs publish demos/run/run.gif)
sed -i -e "s|\(\!\[Run demo\]\).*|\1($RUN_GIF_URL)|" README.md

sleep 1

STATE_GIF_URL=$(vhs publish demos/state/state.gif)
sed -i -e "s|\(\!\[State demo\]\).*|\1($STATE_GIF_URL)|" README.md

sleep 1

TASKS_GIF_URL=$(vhs publish demos/tasks/tasks.gif)
sed -i -e "s|\(\!\[Tasks demo\]\).*|\1($TASKS_GIF_URL)|" README.md
