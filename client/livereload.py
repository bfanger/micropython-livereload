import uos
import sys
import usocket
import uselect

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
        self.poller = uselect.poll()
        self.poller.register(self.socket, uselect.POLLIN)
        self.locked = False

    def readblocks(self, block_num, buf):
        self.detect()
        self.locked = True
        self.socket.write("R" + str(block_num) + "," + str(len(buf)) + "\n")
        self.socket.readinto(buf)
        self.locked = False

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

        self.detect()
        self.locked = True
        self.socket.write("I" + str(op) + "," + str(arg) + "\n")
        response = self.socket.readline()
        self.locked = False
        return eval(response)

    # Detect if there is data availabel for reading on the socket
    # Which would mean a disconnect or reload trigger.
    def detect(self):
        if self.locked:
            return
        events = self.poller.poll(0)
        if not events:
            return
        (socket, event) = events[0]
        self.locked = True
        if event & uselect.POLLIN:
            cmd = socket.readline()
            if cmd == b"LIVERELOAD\n":
                log("Change detected")
                socket.close()
                restart(0)
                return
        if event & uselect.POLLHUP:
            log("Connection lost")
            restart(3)
            return
        self.locked = False


# Connect to the livereload server
def connect(host, port):
    global disk
    log("Connecting to " + host + ":" + str(port))
    disk = Socketdisk(host, port)
    vfs = uos.VfsFat(disk)
    uos.mount(vfs, "/__livereload__")
    uos.chdir("/__livereload__")
    sys.path.insert(0, "/__livereload__/lib")
    sys.path.insert(0, "/__livereload__")


# Connect to a wifi network
# Not needed for unix port
# Not needed for esp8266 port (which remembers wifi settings across reboots)
def wifi(ssid, password):
    import network

    wlan = network.WLAN(network.STA_IF)
    wlan.active(True)
    wlan.connect(ssid, password)


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


# Detect if the code has changed.
# When no or negative interval is given
# The unix port doesn't have Timers and must therefor call `liveload.detect()` directly to enable livereload.
def detect(interval_ms=-1):
    global disk
    global timer
    disk.detect()
    if interval_ms <= 0:
        return
    from machine import Timer

    timer = Timer(-1)
    timer.init(
        period=interval_ms, mode=Timer.PERIODIC, callback=lambda _: disk.detect()
    )


# Restart the program (using the updated source files)
def restart(countdown):
    if countdown:
        import utime

        log("Restarting in:")
        log(3)
        utime.sleep(1)
        log(2)
        utime.sleep(1)
        log(1)
        utime.sleep(1)
    else:
        log("Restarting now...")

    try:
        from machine import reset

        reset()
    except ImportError:
        log("machine.reset() not available")
        sys.exit(808)
