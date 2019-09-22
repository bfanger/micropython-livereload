# Ramdisk implements the uos.AbstractBlockDev interface.
# https://docs.micropython.org/en/latest/library/uos.html#block-devices
class Ramdisk:
    def __init__(self, block_size, num_blocks):
        self.block_size = block_size
        self.data = bytearray(block_size * num_blocks)

    def readblocks(self, block_num, buf):
        for i in range(len(buf)):
            buf[i] = self.data[block_num * self.block_size + i]

    def writeblocks(self, block_num, buf):
        for i in range(len(buf)):
            self.data[block_num * self.block_size + i] = buf[i]

    def ioctl(self, op, arg):
        if op == 4:  # get number of blocks
            return len(self.data) // self.block_size
        if op == 5:  # get block size
            return self.block_size
