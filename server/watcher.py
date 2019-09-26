import uos


class Watcher:
    def __init__(self, folder):
        self.entries = []
        self.mtime = 0
        for entry in uos.listdir(folder):
            path = folder + "/" + entry
            # @todo only watch *.py files?
            self.entries.append(path)
            mtime = uos.stat(path)[-1]
            if mtime > self.mtime:
                self.mtime = mtime

    def detect(self):
        changed = False
        for path in self.entries:
            mtime = uos.stat(path)[-1]
            if mtime > self.mtime:
                changed = True
                print(path, "changed")
                self.mtime = mtime

        return changed

