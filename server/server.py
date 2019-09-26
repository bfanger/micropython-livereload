import usocket
import uselect
import gc


class Server:
    def __init__(self, create_handler):
        self.create_handler = create_handler
        self.socket = usocket.socket(usocket.AF_INET, usocket.SOCK_STREAM)
        self.socket.setsockopt(usocket.SOL_SOCKET, usocket.SO_REUSEADDR, 1)
        self.connections = []
        self.poller = uselect.poll()
        self.poller.register(self.socket, uselect.POLLIN)
        self.idleInterval = 500

    def listen_and_serve(self, host, port):
        try:
            self.listen(host, port)
            self.serve()

        except KeyboardInterrupt:
            pass
        finally:
            self.socket.close()

    def listen(self, host, port):
        addr = usocket.getaddrinfo(host, port)[0][-1]

        self.socket.bind(addr)
        self.socket.listen(0)
        print("Listening on", host, "port", port)

    def serve(self):
        while True:
            ready = self.poller.poll(self.idleInterval)
            if not ready:
                self.idle()
                continue
            for (socket, event) in ready:
                # print(socket, event)
                if socket == self.socket:
                    conn, _ = socket.accept()
                    self.add_connection(conn)
                else:
                    for (conn, handler) in self.connections:
                        if socket == conn:
                            if event & uselect.POLLIN:
                                handler.procesRequest()
                            if event & uselect.POLLHUP:
                                self.remove_connection(conn)
                            if event & uselect.POLLERR:
                                print("poll error")
                                self.remove_connection(conn)

    def idle(self):
        pass

    def add_connection(self, socket):
        self.connections.append((socket, self.create_handler(socket)))
        self.poller.register(socket, uselect.POLLIN)
        print("connections:", len(self.connections))

    def remove_connection(self, socket):
        self.poller.unregister(socket)
        self.connections = [conn for conn in self.connections if conn[0] != socket]
        print("connections:", len(self.connections))

        gc.collect()
