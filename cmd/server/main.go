package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

var (
	dev        *string = flag.String("dev", "/proc/acpi/ibm/fan", "the path to the fan device")
	streamPath *string = flag.String("stream", "stream", "path to a file which is used to receive requests")
)

func init() {
	flag.Parse()

	if *streamPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// if err := cleanStreamFile(); err != nil {
	// 	log.Fatal(err)
	// }
}

func errsAny(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}

	return nil
}

func cleanStreamFile() error {
	err1 := os.Remove(*streamPath)
	file, err2 := os.Create(*streamPath)
	defer file.Close()
	return errsAny(err1, err2)
}

func ApplyLevel(level string) error {
	// echo level 0 | tee /proc/acpi/ibm/fan

	stdin := &bytes.Buffer{}

	tee := exec.Command("tee", *dev)
	tee.Stdin = stdin
	tee.Stderr = os.Stderr
	tee.Stdout = log.Writer()

	stdin.WriteString("level " + level + "\n")

	return tee.Run()
}

var acceptingFanLevels = map[string]struct{}{
	"0":    struct{}{},
	"1":    struct{}{},
	"2":    struct{}{},
	"3":    struct{}{},
	"4":    struct{}{},
	"6":    struct{}{},
	"auto": struct{}{},
}

func logic() error {

	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()

	tail := exec.Command("tail", "-f", *streamPath)
	tail.Stdout = pipeWriter
	tail.Stderr = log.Writer()

	errs := make(chan error, 3)

	go func() {
		errs <- tail.Run()
	}()

	go func() {
		signalC := make(chan os.Signal)
		signal.Notify(signalC, os.Interrupt, os.Kill)
		errs <- fmt.Errorf("terminated with %v", <-signalC)
	}()

	go func() {
		defer func() { errs <- nil }()

		bufferedTail := bufio.NewReader(pipeReader)

		for {
			line, err := bufferedTail.ReadString(byte('\n'))
			if err == io.EOF {
				break
			}

			if err != nil {
				errs <- err
			}

			line = strings.Trim(line, "\n \t")
			log.Printf("got line: %q\n", line)

			_, ok := acceptingFanLevels[line]
			if !ok {
				log.Printf("level not accepted")
				continue
			}

			err = ApplyLevel(line)
			if err != nil {
				errs <- err
			}
		}
	}()

	return <-errs
}

func main() {
	if err := logic(); err != nil {
		log.Println(err)
	}
}
