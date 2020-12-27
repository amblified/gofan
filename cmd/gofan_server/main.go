package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"gitlab.com/malte-L/go_fan/fan"
)

var (
	streamPath *string = flag.String("stream", "", "[required] path to a file which is used to receive requests")
	devPath    *string = flag.String("dev", "", "[required] path to fan device. something like \"/proc/acpi/ibm/fan\"")

	stream io.ReadCloser
)

func requiredString(variable *string) {
	if *variable == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func initStream() (io.ReadCloser, func(chan<- error), error) {
	_, streamFileName := filepath.Split(*streamPath)
	abs, err := filepath.Abs(streamFileName)
	if err != nil {
		return nil, nil, err
	}

	if filepath.Ext(streamFileName) != ".sock" {
		log.Printf("using %q as a text-file for communication\n", abs)

		stream, pipeWriter := io.Pipe()

		tail := exec.Command("tail", "-f", *streamPath)
		tail.Stdout = pipeWriter
		tail.Stderr = log.Writer()

		return stream, func(errs chan<- error) {
			defer pipeWriter.Close()
			errs <- tail.Run()
		}, nil
	}

	log.Printf("using %q as a unix socket for communication", abs)

	socketExists := true

	_, err = os.Stat(*streamPath)
	switch err.(type) {
	case *os.PathError:
		socketExists = false
	}

	l, err := net.Listen("unix", *streamPath)
	if err != nil {
		return nil, nil, err
	}
	if !socketExists {
		// in this case net.Listen() has created the socket
		// but the server runs as sudo, thus not everyone will be
		// able to access that socket
		perm := 0x1FF // TODO: define permissions more precisely
		mode := (int(os.ModeSocket) & ^0x1FF) | perm
		os.Chmod(*streamPath, os.FileMode(mode))
	}

	stream, err := l.Accept()
	if err != nil {
		return nil, nil, err
	}
	log.Printf("successfully established connection")

	return stream, func(chan<- error) {
		l.Close()
	}, nil
}

func init() {
	flag.Parse()

	requiredString(streamPath)
	requiredString(devPath)
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

	s, routine, err := initStream()
	if err != nil {
		return err
	}
	stream = s
	defer stream.Close()

	errs := make(chan error, 3)

	go routine(errs)

	go func() {
		signalC := make(chan os.Signal)
		signal.Notify(signalC, os.Interrupt, os.Kill)
		errs <- fmt.Errorf("terminated with %v", <-signalC)
	}()

	go func() {
		defer func() { errs <- nil }()

		bufferedTail := bufio.NewReader(stream)

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
