package main

import (
	"io"
	"os"
	"os/exec"
)

var NotifyErr = &notifyErr{}

var _ io.Writer = &notifyErr{}

type notifyErr struct {
}

func (n notifyErr) Write(bytes []byte) (int, error) {
	cmd := exec.Command("notify-send", "--urgency=critical", "go_fan: "+string(bytes))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = nil

	return 0, cmd.Run()
}
