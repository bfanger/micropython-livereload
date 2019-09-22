import uos
import sys

import socketdisk


vfs = uos.VfsFat(socketdisk.Socketdisk("10.0.0.20", 60606))
uos.mount(vfs, "/livereload")


uos.chdir("/livereload")

file = open("main.py")
print(file.read())
