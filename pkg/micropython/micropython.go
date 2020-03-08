package micropython

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.bug.st/serial"
)

type Interpreter interface {
	Run(script string) ([]byte, error) // Run script from empty state
	Eval(code string) ([]byte, error)  // Evalute code
	Close() error
}

type CLI struct {
	Command string
	Dir     string
	cmd     *exec.Cmd
}

// Eval: Run python code and return the results (if any)
func (cli *CLI) Run(script string) ([]byte, error) {
	cmd := exec.Command(cli.Command)
	cli.cmd = cmd
	cmd.Dir = cli.Dir
	cmd.Stdin = bytes.NewBufferString(script)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	cli.cmd = nil
	if stderr.Len() != 0 {
		return nil, errors.Errorf("python error: %s", stderr)
	}
	if err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

func (cli *CLI) Close() error {
	if cli.cmd == nil {
		return nil
	}
	return cli.cmd.Process.Kill()
}

func (cli *CLI) Eval(script string) ([]byte, error) {
	return nil, errors.New("Not implemented (yet)")
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
	Running
)

const debug = false

/**
 * Uses the REPL to run python scripts
 * https://docs.micropython.org/en/latest/reference/repl.html
 */

type Board struct {
	io   io.ReadWriteCloser
	mode Mode
	err  error
	out  chan byte
}

func Open(path string, baud int) (*Board, error) {
	io, err := serial.Open(path, &serial.Mode{BaudRate: baud})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't open serial")
	}
	b := &Board{
		io:   io,
		mode: Unknown,
		out:  make(chan byte),
	}
	go func() {
		buf := make([]byte, 1)
		r := bufio.NewReader(io)
		for {
			if _, err := r.Read(buf); err != nil {
				b.err = err
				if debug {
					fmt.Println("read failed")
				}
				close(b.out)
				return
			}
			if debug {
				if int(buf[0]) < 32 {
					fmt.Printf("[%d]", buf[0])
				} else {
					fmt.Print(string(buf))
				}
			}
			b.out <- buf[0]
		}
	}()
	return b, nil
}

func (b *Board) Close() error {
	if b.mode == RawRepl {
		// Send Ctrl+B (Exit raw REPL) for faster REPL detection.
		b.io.Write([]byte{2})
	}
	return b.io.Close()
}

func (b *Board) Run(script string) ([]byte, error) {
	if err := b.Reset(); err != nil {
		return nil, err
	}
	return b.Eval(script)
}
func (b *Board) Eval(code string) ([]byte, error) {
	if err := b.openRawRepl(); err != nil {
		return nil, err
	}
	if _, err := b.io.Write([]byte(code)); err != nil {
		return nil, err
	}
	// Send Ctrl+D (End of Transmission)
	if _, err := b.io.Write([]byte{4}); err != nil {
		return nil, err
	}
	if out, err := b.readUntil([]byte("OK"), 0); err != nil {
		return out, err
	}
	b.mode = Running
	out, err := b.readUntil([]byte{4, '>'}, 0)
	if err != nil {
		return out, err
	}
	b.mode = RawRepl

	if len(out) != 1 {
		if out[0] == 4 { // python error
			return nil, errors.New(string(out[1:]))
		}
	}
	return out, nil
}

func (b *Board) Reset() error {
	if err := b.openRepl(); err != nil {
		return err
	}
	b.io.Write([]byte{4})
	b.mode = Unknown

	return b.openRepl()
}
func (b *Board) HardReset() error {
	b.Eval(`
import machine
machine.reset()
`)
	b.mode = Unknown
	return nil
}

func (b *Board) openRawRepl() error {
	if b.mode == RawRepl {
		return nil
	}
	if err := b.openRepl(); err != nil {
		return errors.Wrap(err, "could open REPL")
	}
	// Send Ctrl+A (Start of Heading) to open raw REPL
	if _, err := b.io.Write([]byte{1}); err != nil {
		return err
	}
	if _, err := b.readUntil([]byte{'>'}, 100*time.Millisecond); err != nil {
		return errors.Wrap(err, "could open raw mode")
	}
	b.mode = RawRepl
	return nil
}

func (b *Board) openRepl() error {
	prompt := []byte{'>', '>', '>', 32}
	switch b.mode {
	case Repl:
		return nil
	case RawRepl:
		b.io.Write([]byte{2}) // Ctrl+B (Exit raw REPL)
		if _, err := b.readUntil(prompt, 250*time.Millisecond); err != nil {
			return err
		}
		b.mode = Repl
		return nil
	case Unknown:
		_, err := b.readUntil(prompt, time.Millisecond) // prompt in buffer?
		if err != nil {
			if _, err := b.io.Write([]byte{3}); err != nil { // Ctrl+C stop running script
				return err
			}
			_, err = b.readUntil(prompt, 1500*time.Millisecond)
			if err != nil {
				if _, err := b.io.Write([]byte{2}); err != nil { // Ctrl+B Exit raw or reboot
					return err
				}
				if _, err = b.readUntil(prompt, 5*time.Second); err != nil {
					return err
				}
			}
		}
		b.mode = Repl
		return nil

	default:
		return errors.Errorf("unexpected mode: %d", b.mode)
	}
}

func (b *Board) readUntil(sequence []byte, d time.Duration) ([]byte, error) {
	if b.err != nil {
		return nil, b.err
	}
	var out bytes.Buffer
	buf := make([]byte, len(sequence))
	for {
		timeout := time.After(d)
		if d == 0 {
			timeout = nil
		}
		select {
		case <-timeout:
			return nil, errors.Errorf("timed out %s", d)
		case char, ok := <-b.out:
			if ok == false {
				return nil, b.err
			}
			out.WriteByte(char)
			buf = append(buf[1:], char)
			if bytes.Compare(buf, sequence) == 0 {
				return out.Bytes()[:out.Len()-len(sequence)], nil
			}
		}
	}
}
