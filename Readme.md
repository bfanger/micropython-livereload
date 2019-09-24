# LiveReload for Microcontrollers

Execute micropython code on a microcontroller directly from a folder on your PC.

## Setup

The server need a micropython with VfsFat enabled.
(Not the regular unix port, but included in the micropython_coverage build)

### Via docker

```
docker build --tag micropython-coverage .
alias mpy="docker run -v $PWD:/app --rm -it micropython-coverage"
```

### macOS

Download full source from: http://www.micropython.org/download

```
brew install libffi
export PKG_CONFIG_PATH="/usr/local/opt/libffi/lib/pkgconfig"

cd micropython-?.??/ports/unix
make coverage
```

# How it works

When microcontroller connects to PC (via socket) the server (PC) creates a new blockdevice in-memory, formats it in FAT and copies the contents of the given folder into that blockdevice. The microcontroller (client) mounts that blockdevice (via socket) and import the scripts from there

# Idea's / Goals / Roadmap

- Allow for multiple microcontrollers (running micropython) to be connected to a single pc.
- Automaticly restart when a python file changed.
- Option to use mpy-cross to pre-compile the python scripts to bytecode.
