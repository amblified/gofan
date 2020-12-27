package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/malte-L/go_fan/fan"
)

var (
	streamPath  *string = flag.String("stream", "", "[required] path to a file which is used to receive requests")
	rulesetPath *string = flag.String("rules", "", "[required] path to a file which stores the ruleset")
	devPath     *string = flag.String("dev", "", "[required] path to fan device. something like \"/proc/acpi/ibm/fan\"")

	stream *os.File
)

func requiredString(variable *string) {
	if *variable == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func init() {
	flag.Parse()

	requiredString(streamPath)
	requiredString(rulesetPath)
	requiredString(devPath)

	var err error
	stream, err = os.OpenFile(*streamPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatal(err)
	}
}

func sendLevelChangeRequest(level string) error {
	_, err := stream.WriteString(level + "\n")
	return err
}

func logic(ruleset *fan.Ruleset) error {

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
	defer stream.Close()

	ruleset, err := fan.ReadRuleset(*rulesetPath)
	if err != nil {
		log.Fatalf("could not read ruleset")
	}

	if err := logic(ruleset); err != nil {
		fmt.Fprintf(fan.NotifyErr, "error occured")
		log.Fatal(err)
	}
}
