import unittest
import multiplex
import io


class TestMultiplex(unittest.TestCase):
    def test_read(self):
        message = multiplex.read(io.BytesIO(b"<MSG 2 A>OK</MSG>"))

        self.assertEqual(message, ("A", "OK", 14))


if __name__ == "__main__":
    unittest.main()
