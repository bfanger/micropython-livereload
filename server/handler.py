class Handler:
    def __init__(self, disk, conn):
        self.disk = disk
        self.conn = conn

    def handle(self):
        while True:
            try:
                line = str(self.conn.readline(), "ascii")
                args = line[1:-1].split(",")
                if not line:
                    print("Disconnected?", line)
                    break
                if line[0] == "I":
                    print("IOCTL", args)
                    val = self.disk.ioctl(int(args[0]), int(args[1]))
                    self.conn.send(encode(val) + "\n")
                    continue
                if line[0] == "R":
                    print("READ", args)
                    buf = bytearray(int(args[1]))
                    self.disk.readblocks(int(args[0]), buf)
                    self.conn.send(buf)
                    continue

                print("Unknown command", line)
                # self.conn.send("O1")
            except OSError:
                print("connection error")
                return False
        return True


def encode(value):
    if value is None:
        return "None"
    if isinstance(value, int):
        return str(value)
    raise Exception("No encoder for ", type(value))
