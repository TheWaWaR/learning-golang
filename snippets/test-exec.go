package main

import (
	"fmt"
	"flag"
	"os/exec"
)

var path = flag.String("path", ".", "Path to `du`")

func init() {
	flag.Parse()
}

type MyCmd struct {
	name string
	args []string
	env  []string
}

func main() {
	cmd0 := MyCmd{"du", []string{"--max-depth=1", "-h", *path}, []string{"ABC=1"}}
	cmd1 := MyCmd{"du", []string{"--max-depth=1", "-h", *path}, []string{}}
	cmd2 := MyCmd{"du", []string{"--max-depth=2", "-h", *path}, []string{}}
	cmd3 := MyCmd{"du", []string{"--max-depth=1", "-h", *path}, []string{}}
	cmd4 := MyCmd{"du", []string{"--max-depth=1", "-h", *path}, []string{}}

	for _, c := range []MyCmd{cmd0, cmd1, cmd2, cmd3, cmd4} {
		out, _ := exec.Command(c.name, c.args...).Output()
		fmt.Printf("Output:\n------\n%s\n", out)
	}
}
