import usocket

# Socketdisk implements the uos.AbstractBlockDev interface.
# https://docs.micropython.org/en/latest/library/uos.html#block-devices
class Socketdisk:
    def __init__(self, host, port):
        self.addr = usocket.getaddrinfo(host, port)[0][-1]
        self.conn = usocket.socket(usocket.AF_INET, usocket.SOCK_STREAM)

    def readblocks(self, block_num, buf):
        self.conn.send("R" + str(block_num) + "," + str(len(buf)) + "\n")
        self.conn.readinto(buf)

    def writeblocks(self, block_num, buf):
        raise Exception("readonly (for now)")

    def ioctl(self, op, arg):
        # print("op", op, arg)
        if op == 1:  # initialise the device
            self.conn.connect(self.addr)
            return
        elif op == 2:  # shutdown the device
            self.conn.close()
            return

        self.conn.send("I" + str(op) + "," + str(arg) + "\n")
        response = self.conn.readline()
        return eval(response)
