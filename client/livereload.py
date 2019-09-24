import uos
import sys
import usocket

DEBUG = True


def log(*message):
    if DEBUG:
        print("[\033[1;36mlivereload\033[0;0m]", *message)


# Socketdisk implements the uos.AbstractBlockDev interface via socket.
# https://docs.micropython.org/en/latest/library/uos.html#block-devices
# https://docs.micropython.org/en/latest/library/usocket.html
class Socketdisk:
    def __init__(self, host, port):
        self.addr = usocket.getaddrinfo(host, port)[0][-1]
        self.socket = usocket.socket(usocket.AF_INET, usocket.SOCK_STREAM)

    def readblocks(self, block_num, buf):
        self.socket.write("R" + str(block_num) + "," + str(len(buf)) + "\n")
        self.socket.readinto(buf)

    def writeblocks(self, block_num, buf):
        raise Exception("readonly (for now)")

    def ioctl(self, op, arg):
        if op == 1:  # 1: initialise the device
            if hasattr(self.socket, "timeout"):
                self.socket.timeout(15)
            self.socket.connect(self.addr)
            return
        elif op == 2:  # 2: shutdown the device
            self.socket.close()
            return

        self.socket.write("I" + str(op) + "," + str(arg) + "\n")
        response = self.socket.readline()
        return eval(response)


def connect(host, port):
    log("Connecting to " + host + ":" + str(port))
    vfs = uos.VfsFat(Socketdisk(host, port))
    uos.mount(vfs, "/__livereload__")

    uos.chdir("/__livereload__")
    sys.path.insert(0, "/__livereload__/lib")
    sys.path.insert(0, "/__livereload__")


# Wait until the wifi is connected
def wait_for_network():
    try:
        import network
        import machine
    except ImportError:
        # Assume connection (unix port)
        return

    log("Waiting for network...")
    wlan = network.WLAN(network.STA_IF)
    while not wlan.isconnected():
        status = wlan.status()
        if status == network.STAT_CONNECTING:
            machine.idle()
        else:
            log("Network status:", status)
            break
