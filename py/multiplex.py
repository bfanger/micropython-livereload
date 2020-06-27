# uart.read(10)       # read 10 characters, returns a bytes object
# uart.read()         # read all available characters
# uart.readline()     # read a line
# uart.readinto(buf)  # read and store into the given buffer
# uart.write('abc')


__ESCAPE_CHAR = b"\\"
__OPEN_TAG = b"<MSG "
__CLOSE_TAG = b"</MSG>"
__ESCAPED_TAG = __ESCAPE_CHAR + __OPEN_TAG
__MIN_SIZE = len(__OPEN_TAG) + 3 + len(__CLOSE_TAG)

__MODE_OPEN = 1
__MODE_LENGTH = 2
__MODE_CHANNEL = 3
__MODE_END = 4

# Read a message from a stream
# Returns a tuple of channel, body and the number of bytes read
def read(stream):
    cursor = 0
    pos = 0
    mode = __MODE_OPEN
    char = b" "
    buffer = b""
    while True:
        prev = char
        char = stream.read(1)
        if not char:
            break
        pos += 1
        if mode == __MODE_OPEN:
            if cursor == 0 and prev == __ESCAPED_TAG:
                continue
            if char != __OPEN_TAG[cursor]:
                if char == __OPEN_TAG[0]:
                    cursor = 1
                else:
                    cursor = 0
                continue

            cursor += 1
            if cursor > len(__OPEN_TAG) - 1:
                mode = __MODE_LENGTH
                continue

        if mode == __MODE_LENGTH:
            if char != b" ":
                buffer += char
            else:
                length = int(buffer)
                mode = __MODE_CHANNEL
                buffer = b""
                continue

        if mode == __MODE_CHANNEL:
            if char != ">":
                buffer += char
                continue
            channel = buffer.decode("utf-8")
            # mode BODY
            body = stream.read(length).replace(__ESCAPED_TAG, __OPEN_TAG)
            mode = __MODE_END
            cursor = 0
            continue

        if mode == __MODE_END:
            if char != __CLOSE_TAG[cursor]:
                raise Exception("Missing closing tag")

            cursor += 1
            if cursor > len(__OPEN_TAG) - 1:
                return (channel, body, pos)

