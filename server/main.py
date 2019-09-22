import sys
import server

if len(sys.argv) != 2:
    print("Usage: " + sys.argv[0] + " [path]")
else:
    s = server.Server(sys.argv[1])
    s.listen_and_serve("0.0.0.0", 60606)
