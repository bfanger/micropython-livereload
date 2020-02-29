from uos import VfsFat, mount
from binascii import b2a_base64, hexlify
from ramdisk import Ramdisk
from sys import stdout


class Filesystem:
    def __init__(self, disk, mnt):
        self.disk = disk
        self.mnt = mnt
        VfsFat.mkfs(disk)
        self.vfs = VfsFat(disk)
        mount(self.vfs, mnt)

    def dump(self):
        output = hexlify(self.disk.data)
        stdout.write(output)

    def add(self, src, dest):
        # @todo auto create folders + dest optional
        srcfile = open(src, "r")
        destfile = open(self.mnt + "/" + dest, "w")
        destfile.write(srcfile.read())
        srcfile.close()
        destfile.close()


def create(kb):
    disk = Ramdisk(512, kb * 2)
    return Filesystem(disk, "/ramdisk")

