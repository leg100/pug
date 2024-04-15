vhs demos/modules.tape
MODULES_GIF_URL=$(vhs publish demos/modules.gif)
sed -i -e "s|\(\!\[Modules demo\]\).*|\1($MODULES_GIF_URL)|" README.md
