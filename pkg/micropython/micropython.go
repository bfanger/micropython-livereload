package micropython

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/term"
)

type Interpreter interface {
	Eval(code string) ([]byte, error)
}

type CLI struct {
	Command string
	Dir     string
}

// Eval: Run python code and return the results (if any)
func (cli *CLI) Eval(code string) ([]byte, error) {
	cmd := exec.Command(cli.Command)
	cmd.Dir = cli.Dir
	cmd.Stdin = bytes.NewBufferString(code)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if stderr.Len() != 0 {
		return nil, errors.Errorf("python error: %s", stderr)
	}
	if err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

type Info struct {
	Name    string
	Version string
}

func GetInfo(i Interpreter) (*Info, error) {
	output, err := i.Eval(`
import sys;
print(sys.implementation.name)
version = sys.implementation.version
if len(version) == 3:
	print("%d.%d.%d" % version)
else:
	print("%d.%d.%d" % (version.major, version.minor, version.micro))
`)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	return &Info{
		Name:    lines[0],
		Version: lines[1],
	}, err
}

type Board struct {
	tty *term.Term
	// port   serial.Port
	Out *bufio.Reader
	// Input *bufio.Writer
}

func Open(path string, baud int) (*Board, error) {
	// port, err := serial.Open(path, &serial.Mode{BaudRate: baud})
	tty, err := term.Open(path, term.Speed(baud), term.RawMode)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't open serial")
	}
	return &Board{
		tty: tty,
		Out: bufio.NewReader(tty),
		// Input: bufio.NewWriter(tty),
	}, nil

}
func (b *Board) Close() error {
	if err := b.tty.Restore(); err != nil {
		return err
	}
	return b.tty.Close()
}
func (b *Board) Eval(code string) ([]byte, error) {
	script := []byte(code)
	ctrlc := []byte{3}
	b.tty.Write(ctrlc)
	var buf bytes.Buffer
	gt := 0
	reading := false
	for {
		char := make([]byte, 1)
		if _, err := b.Out.Read(char); err != nil {
			return nil, err
		}
		if reading {
			buf.Write(char)
		}
		if char[0] == 32 {
			if gt == 3 {
				if reading {
					return buf.Bytes()[len(script)+2 : buf.Len()-4], nil
				}
				b.tty.Write(script)
				b.tty.Write([]byte{13})
				reading = true
			}
		}
		if char[0] == '>' {
			gt++
		} else {
			gt = 0
		}
	}
}

// ReadUntil : Read until there is no new data for x duration
func ReadUntil(r io.Reader, d time.Duration) ([]byte, error) {
	// var output bytes.Buffer
	// errs := make(chan error)

	// pr, pw := io.Pipe()

	//
	// var m sync.Mutex
	// done := false
	// go func() {
	// 	buf := make([]byte, 1024)
	// 	for {
	// 		l, err := r.Read(buf)
	// 		if err != nil {
	// 			errs <- err
	// 			return
	// 		}
	// 		// fmt.Printf("%s", buf[0:l]) // works?
	// 		m.Lock()
	// 		if done {
	// 			m.Unlock()
	// 			return
	// 		}
	// 		w.Write(buf[0:l])
	// 		m.Unlock()
	// 	}
	// }()
	// for {
	// 	select {
	// 	case err := <-errs:
	// 		return nil, err
	// 	case <-time.After(d):
	// 		// m.Lock()
	// 		// fmt.Println(w.Buffered())
	// 		// b2 := make([]byte, 50)
	// 		// pr.Read(b2)
	// 		// _ = pr
	// 		// if output.Len() != 0 {
	// 		// 	done = true

	// 		// 	m.Unlock()
	// 		// 	return output.Bytes(), nil
	// 		// }
	// 		// m.Unlock()
	// 	}
	// }
	return nil, nil
}
