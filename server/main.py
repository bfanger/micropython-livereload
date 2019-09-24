import sys
from server import Server

if len(sys.argv) != 2:
    print("Usage: " + sys.argv[0] + " [path]")
else:
    s = Server(sys.argv[1])
    s.listen_and_serve("0.0.0.0", 1808)
