package micropython

import (
	"bytes"
	"os/exec"

	"github.com/pkg/errors"
)

type Interpreter struct {
	Command string
	Dir     string
}

// Eval: Run python code and return the results (if any)
func (i *Interpreter) Eval(code string) ([]byte, error) {
	cmd := exec.Command(i.Command)
	cmd.Dir = i.Dir
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
