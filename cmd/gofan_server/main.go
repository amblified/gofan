package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"gitlab.com/malte-L/go_fan/fan"
)

var (
	streamPath *string = flag.String("stream", "", "[required] path to a file which is used to receive requests")
	devPath    *string = flag.String("dev", "", "[required] path to fan device. something like \"/proc/acpi/ibm/fan\"")
)

func requiredString(variable *string) {
	if *variable == "" {
		flag.Usage()
		os.Exit(1)
	}
}

// not working as intended yet... use unix sockets instead
func initStreamFile() {
	_ = os.Remove(*streamPath)
	file, err := os.Create(*streamPath)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
}

func init() {
	flag.Parse()

	requiredString(streamPath)
	requiredString(devPath)

	// initStreamFile()
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

			err = fan.ApplyLevel(*devPath, line)
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
