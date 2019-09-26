import uos

# import sys
# import gc
from ramdisk import Ramdisk


class Handler:
    def __init__(self, socket, folder):
        self.socket = socket
        self.folder = folder
        self._create_disk()

    def _create_disk(self):
        mnt = "/ramdisk"
        self.disk = Ramdisk(512, 200)  # 500 KiB
        uos.VfsFat.mkfs(self.disk)
        vfs = uos.VfsFat(self.disk)
        uos.mount(vfs, mnt)

        entries = uos.listdir(self.folder)
        self.mtime = 0
        for entry in entries:
            path = self.folder + "/" + entry
            mtime_entry = uos.stat(path)[-1]
            if self.mtime < mtime_entry:
                self.mtime = mtime_entry
            # print(entry, mtime_entry)
            src = open(path, "r")
            dest = open(mnt + "/" + entry, "w")
            dest.write(src.read())
            src.close()
            dest.close()
        uos.umount(mnt)

    def procesRequest(self):
        try:
            line = str(self.socket.readline(), "ascii")
            args = line[1:-1].split(",")
            if not line:  # HUP
                return False
            if line[0] == "I":
                # print("IOCTL", args)
                val = self.disk.ioctl(int(args[0]), int(args[1]))
                self.socket.write(encode(val) + "\n")
                return True
            if line[0] == "R":
                # print("READ", args)
                buf = bytearray(int(args[1]))
                self.disk.readblocks(int(args[0]), buf)
                self.socket.write(buf)
                return True

            print("Unknown command", line)
        except OSError:
            print("connection error")
        return False

    def livereload(self):
        self.socket.write("LIVERELOAD\n")


def encode(value):
    if value is None:
        return "None"
    if isinstance(value, int):
        return str(value)
    raise Exception("No encoder for ", type(value))

