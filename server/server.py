import uos
import usocket

from ramdisk import Ramdisk
from handler import Handler


class Server:
    def __init__(self, folder):
        self._counter = 0
        self.folder = folder

    def _create_disk(self):
        self._counter += 1
        target = "/ramdisk" + str(self._counter)
        disk = Ramdisk(512, 500)  # 250 KiB
        uos.VfsFat.mkfs(disk)
        vfs = uos.VfsFat(disk)
        uos.mount(vfs, target)

        entries = uos.listdir(self.folder)
        for entry in entries:
            print(entry)
            src = open(self.folder + "/" + entry, "r")
            dest = open(target + "/" + entry, "w")
            dest.write(src.read())
            src.close()
            dest.close()
        uos.umount(target)
        return disk

    def listen_and_serve(self, host, port):
        try:
            addr = usocket.getaddrinfo(host, port)[0][-1]
            socket = usocket.socket(usocket.AF_INET, usocket.SOCK_STREAM)
            socket.setsockopt(usocket.SOL_SOCKET, usocket.SO_REUSEADDR, 1)
            socket.bind(addr)
            socket.listen(0)
            print("Listening on", host, "port", port)
            while True:
                print("Waiting for client")
                conn, addr = socket.accept()
                print("Connected")
                disk = self._create_disk()
                handler = Handler(disk, conn)
                handler.handle()

        except KeyboardInterrupt:
            pass
        finally:
            socket.close()

