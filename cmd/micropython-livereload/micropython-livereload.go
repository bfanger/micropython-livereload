package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/bfanger/micropython-livereload/pkg/micropython"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var cmd string
	flag.StringVar(&cmd, "micropython", "", "Path to the coverage variant of micropython")
	flag.Parse()
	if cmd == "" {
		cmd = cwd + "/micropython-coverage"
	}
	mpy := &micropython.Interpreter{
		Command: cmd,
		Dir:     cwd + "/py",
	}
	output, err := mpy.Eval(`
import fatfs;
d = fatfs.create(50)
d.add("../example/main.py", "main.py")
d.dump()
`)
	if err != nil {
		panic(err)
	}
	ram := make([]byte, hex.DecodedLen(len(output)))
	if _, err = hex.Decode(ram, output); err != nil {
		fmt.Printf("%s", output)
		panic(err)
	}
	fmt.Printf("\nsize %dkb\n", len(ram)/1024)
	fmt.Printf("%s", ram)

	// d.add("../example/test.py", "test.py")

}
func calculator(mpy *micropython.Interpreter) {
	code := "print(1 + 2)"
	output, err := mpy.Eval(code)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: \"%s\"", code, output)
}
