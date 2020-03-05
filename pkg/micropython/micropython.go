package micropython

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

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
		Name:    strings.Trim(lines[0], "\r\n"),
		Version: strings.Trim(lines[1], "\r\n"),
	}, err
}

type Mode int

const (
	Unknown Mode = iota
	Repl
	RawRepl
	Output
)

type Board struct {
	tty  *term.Term
	mode Mode
	Out  *bufio.Reader
}

func Open(path string, baud int) (*Board, error) {
	// port, err := serial.Open(path, &serial.Mode{BaudRate: baud})
	tty, err := term.Open(path, term.Speed(baud)) //, term.RawMode
	if err != nil {
		return nil, errors.Wrap(err, "couldn't open serial")
	}
	return &Board{
		tty:  tty,
		Out:  bufio.NewReader(tty),
		mode: Unknown,
	}, nil

}
func (b *Board) Close() error {
	if err := b.tty.Restore(); err != nil {
		return err
	}
	return b.tty.Close()
}

func (b *Board) Eval(code string) ([]byte, error) {
	ctrla := []byte{1}
	ctrlb := []byte{2}
	ctrlc := []byte{3}
	ctrld := []byte{4}
	if b.mode == Unknown {
		b.tty.Write([]byte{13})
		b.tty.Write(ctrlc)
		b.tty.Write(ctrlb)
	} else if b.mode == RawRepl {
		b.tty.Write([]byte(code))
		b.tty.Write(ctrld)
	} else {
		return nil, errors.New("Unexpected mode")
	}
	buf := make([]byte, 4)
	var output bytes.Buffer
	prompt := []byte{'>', '>', '>', 32}
	ok := []byte{'O', 'K'}
	end := []byte{4, 4, '>'}
	for {
		char := make([]byte, 1)
		if _, err := b.Out.Read(char); err != nil {
			return nil, err
		}
		buf = append(buf[1:], char[0])
		// fmt.Print(string(char[0]))

		switch b.mode {
		case Unknown:
			if bytes.Compare(buf, prompt) == 0 {
				b.mode = Repl
				b.tty.Write(ctrla)

			}
		case Repl:
			if char[0] == '>' {
				b.mode = RawRepl
				b.tty.Write([]byte(code))
				b.tty.Write(ctrld)
			}
		case RawRepl:
			if bytes.Compare(buf[2:], ok) == 0 {
				b.mode = Output
			}
		case Output:
			output.Write(char)
			if bytes.Compare(buf[1:], end) == 0 {
				b.mode = RawRepl
				return output.Bytes()[:output.Len()-3], nil
			}
		}
	}
}
