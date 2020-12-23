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

func checkIfShouldDowngrade(currentMode **Mode, ticker *time.Ticker, callback chan<- error) {
	for range ticker.C {
		if *currentMode == nil {
			continue
		}

		temp, err := getTemp()
		if err != nil {
			callback <- err
			return
		}

		if temp > float32((*currentMode).TransitionWhenBelow) {
			continue
		}

		log.Printf("should downgrade\n")
		callback <- nil
	}
}

func checkIfShouldUpgrade(currentMode **Mode, ruleset *Ruleset, ticker *time.Ticker, callback chan<- error) {
	for range ticker.C {
		if *currentMode == nil {
			continue
		}

		temp, err := getTemp()
		if err != nil {
			callback <- err
			return
		}

		mode, err := ruleset.findAppropriateMode(temp)
		notreached(err)

		if mode.Name == (*currentMode).Name {
			continue
		}

		log.Printf("should upgrade (checked with temperature: %f°C)\n", temp)
		callback <- nil
	}
}

// sometimes my os decides it's time for lift-off; even when the current cpu temperatur (and Elon Musk) says it's not. This occurs often when AC is plugged in
// in this case the current level as given by reading from the fan device does not equal the level of the mode we're currently in. In this case we should trigger a recheck on the currently best suited mode
func checkForUnmonitoredDeviceChanges(currentMode **Mode, ticker *time.Ticker, callback chan<- error) {
	for range ticker.C {
		if *currentMode == nil {
			continue
		}

		level, err := GetCurrentFanLevel(dev)
		if err != nil {
			callback <- err
			return
		}

		if level == (*currentMode).Level {
			continue
		}

		log.Printf("umonitored device change\n")
		callback <- nil
	}
}

func checkForTimeout(ruleset Ruleset, ticker *time.Ticker, callback chan<- error) {
	for range ticker.C {
		log.Printf("timed out after %v\n", ruleset.Timeouts.Standard.Duration)
		callback <- nil
	}
}

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
		temp, err := getTemp()
		if err != nil {
			return err
		}
		log.Printf("found: temperature is at %f°C\n", temp)

		log.Printf("searching mode..")
		mode, err = ruleset.findAppropriateMode(temp)
		notreached(err)
		log.Printf("found %q\n", mode.Name)

		log.Printf("applying level %q..", mode.Level)
		err = applyLevel(mode.Level)
		if err != nil {
			return err
		}
		log.Printf("success\n")
		log.Printf("\n")

		err = <-callback

		// select {
		// case err = <-checkIfShouldDowngrade(float32(mode.TransitionWhenBelow), ruleset.Timeouts.Downgrade.Duration):
		// 	log.Printf("should downgrade\n")
		// case err = <-checkIfShouldUpgrade(ruleset, mode.Name, ruleset.Timeouts.Upgrade.Duration):
		// 	log.Printf("should upgrade\n")
		// case err = <-checkForUnmonitoredDeviceChanges(mode.Level, ruleset.Timeouts.UnmonitoredChange.Duration):
		// 	log.Printf("umonitored device change\n")
		// case <-time.After(ruleset.Timeouts.Standard.Duration):
		// 	log.Printf("timed out after %v\n", ruleset.Timeouts.Standard.Duration)
		// }

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
