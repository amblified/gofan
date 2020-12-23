package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

func GetCurrentFanLevel(dev string) (string, error) {
	stdout := &bytes.Buffer{}

	cat := exec.Command("cat", dev)
	cat.Stdout = stdout
	cat.Stderr = os.Stderr
	err := cat.Run()
	if err != nil {
		return "", err
	}

	return parseFanLevel(stdout.String()), nil
}

func parseFanLevel(str string) string {
	const whitespace = " \t"

	lines := strings.Split(str, "\n")

	line := ""
	for _, l := range lines {
		l = strings.TrimLeft(l, whitespace)

		if strings.HasPrefix(l, "level") {
			line = l
			break
		}
	}

	jumpOver := "level:"
	line = line[len(jumpOver):]

	line = strings.TrimLeft(line, whitespace)
	// line = strings.TrimRight(line, whitespace)

	return line
}
