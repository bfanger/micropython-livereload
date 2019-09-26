# LiveReload for Microcontrollers

- Execute python on a microcontroller directly from a folder on your PC.
- Restart when a file has changed.
- Supports multiple microcontrollers at once

## Usage

Edit `client/boot.py` with your IP and WIFI settings
Upload `client/boot.py` and `client/livereload.py` into your microcontroller (using [rshell](https://github.com/dhylands/rshell) or [ampy](https://github.com/pycampers/ampy)

Start the server using the micropython_coverage version

```sh
micropython_coverage server/main.py your_project_folder
```

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

1. The microcontroller connects to server on the PC (via socket).
2. The server (PC) creates a new blockdevice in-memory, formats it in FAT and copies the contents of the given folder into that blockdevice.
3. The client (microcontroller) mounts that blockdevice (via socket) and import the scripts from there.
4. When a file's modified date is changed the server notifies the client and the client reboot.

Every client get it's own immutable filesystem (readonly and remains the same unit the microcontroller reboots).

# Idea's / Goals / Roadmap

- Option to use mpy-cross to pre-compile the python scripts to bytecode.
- Faster reload: Use "soft reboot" (expose `soft_reset()` to python)
- Faster reload: Use usb-serial / UART (skip networking)
- Publish as a upip package

# Known bugs

- Reload on all file saves. Should only reload when a (used) **python** file has **changed**.
- A hard reset is not detected by the socket.
