import sys
from server import Server
from watcher import Watcher
from handler import Handler


if len(sys.argv) != 2:
    print("Usage: " + sys.argv[0] + " [path]")
else:
    folder = sys.argv[1]
    w = Watcher(folder)
    s = Server(lambda socket: Handler(socket, folder))

    def reload():
        if w.detect():
            for (_, handler) in s.connections:
                if w.mtime > handler.mtime:
                    handler.livereload()

    s.idle = reload
    s.listen_and_serve("0.0.0.0", 1808)
