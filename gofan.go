package main

import (
	"fmt"
	"log"
	"time"
)

const (
	dev            = "/proc/acpi/ibm/fan"
	configFileName = "gofan"
)

func logic(ruleset *Ruleset) error {

	var (
		mode     *Mode
		callback = make(chan error)
	)

	go checkForUnmonitoredDeviceChanges(&mode, time.NewTicker(ruleset.Timeouts.UnmonitoredChange.Duration), callback)
	go checkIfShouldUpgrade(&mode, ruleset, time.NewTicker(ruleset.Timeouts.Upgrade.Duration), callback)
	go checkIfShouldDowngrade(&mode, time.NewTicker(ruleset.Timeouts.Downgrade.Duration), callback)
	go checkForTimeout(*ruleset, time.NewTicker(ruleset.Timeouts.Standard.Duration), callback)

	for {
		log.Printf("get temperatur information..")
		temp, err := GetTemp()
		if err != nil {
			return err
		}
		log.Printf("found: temperature is at %f°C\n", temp)

		log.Printf("searching mode..")
		mode, err = ruleset.FindAppropriateMode(temp)
		notreached(err)
		log.Printf("found %q\n", mode.Name)

		log.Printf("applying level %q..", mode.Level)
		err = ApplyLevel(mode.Level)
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
	if err := logic(DefaultRuleset); err != nil {
		log.Fatal(err)
		fmt.Fprintf(NotifyErr, "error occured")
	}
}
