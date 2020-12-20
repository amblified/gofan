package main

import (
	"io"
	"os"
	"os/exec"
)

var Notify = &notify{}

var _ io.Writer = &notify{}

type notify struct {
}

func (n notify) Write(bytes []byte) (int, error) {
	cmd := exec.Command("notify-send", "go_fan: "+string(bytes))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = nil

	return 0, cmd.Run()
}
