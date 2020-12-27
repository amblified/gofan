package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/malte-L/go_fan/fan"
	"golang.org/x/sync/errgroup"
)

const (
	configDirName         = "gofan"
	defaultConfigFileName = "rules"
)

var (
	streamPath  *string = flag.String("stream", "", "[required] path to a file which is used to receive requests")
	rulesetPath *string = flag.String("rules", "", "[required] path to a file which stores the ruleset")
	devPath     *string = flag.String("dev", "", "[required] path to fan device. something like \"/proc/acpi/ibm/fan\"")

	stream io.WriteCloser
)

func requiredString(variable *string) {
	if *variable == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func initStream() error {
	_, streamFileName := filepath.Split(*streamPath)
	abs, err := filepath.Abs(streamFileName)
	if err != nil {
		return err
	}

	if filepath.Ext(streamFileName) != ".sock" {
		log.Printf("using %q as a text-file for communication\n", abs)
		stream, err = os.OpenFile(*streamPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

		return nil
	}

	log.Printf("using %q as a unix socket for communication", abs)
	stream, err = net.Dial("unix", *streamPath)
	if err != nil {
		return err
	}
	log.Printf("successfully established connection")

	return nil
}

func init() {
	flag.Parse()

	requiredString(streamPath)
	requiredString(devPath)

	group := &errgroup.Group{}

	group.Go(func() error {
		var err error
		if *rulesetPath == "" {
			*rulesetPath, err = os.UserConfigDir()
			*rulesetPath = filepath.Join(*rulesetPath, configDirName, defaultConfigFileName)
		}
		return err
	})

	group.Go(initStream)

	if err := group.Wait(); err != nil {
		log.Fatal(err)
	}
}

func sendLevelChangeRequest(level string) error {
	_, err := stream.Write([]byte(level + "\n"))
	return err
}

func logic(ruleset *fan.Ruleset) error {
	defer stream.Close()

	var (
		mode     *fan.Mode
		callback = make(chan error)
	)

	go checkForUnmonitoredDeviceChanges(&mode, time.NewTicker(ruleset.Timeouts.UnmonitoredChange.Duration), callback)
	go checkIfShouldUpgrade(&mode, ruleset, time.NewTicker(ruleset.Timeouts.Upgrade.Duration), callback)
	go checkIfShouldDowngrade(&mode, time.NewTicker(ruleset.Timeouts.Downgrade.Duration), callback)
	go checkForTimeout(*ruleset, time.NewTicker(ruleset.Timeouts.Standard.Duration), callback)

	for {
		log.Printf("get temperatur information..")
		temp, err := fan.GetTemp()
		if err != nil {
			return err
		}
		log.Printf("found: temperature is at %f°C\n", temp)

		log.Printf("searching mode..")
		mode, err = ruleset.FindAppropriateMode(temp)
		notreached(err)
		log.Printf("found %q\n", mode.Name)

		log.Printf("applying level %q..", mode.Level)
		err = sendLevelChangeRequest(mode.Level)
		if err != nil {
			return err
		}
		log.Printf("success\n")
		log.Printf("\n")

		err = <-callback

		if err != nil {
			return err
		}

	}
}

func main() {
	log.Printf("go\n")

	ruleset, err := fan.ReadRuleset(*rulesetPath)
	if err != nil {
		log.Fatalf("could not read ruleset from %q", *rulesetPath)
	}

	if err := logic(ruleset); err != nil {
		fmt.Fprintf(fan.NotifyErr, "error occured")
		log.Fatal(err)
	}
}
