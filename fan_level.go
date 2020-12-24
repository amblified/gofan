package main

import (
	"bytes"
	"log"
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

	return ParseFanLevel(stdout.String()), nil
}

func ParseFanLevel(str string) string {
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

func ApplyLevel(level string) error {
	// echo level 0 | tee /proc/acpi/ibm/fan

	stdin := &bytes.Buffer{}

	tee := exec.Command("tee", dev)
	tee.Stdin = stdin
	tee.Stderr = os.Stderr
	tee.Stdout = log.Writer()

	stdin.WriteString("level " + level + "\n")

	return tee.Run()
}
