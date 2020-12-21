package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	dev            = "/proc/acpi/ibm/fan"
	configFileName = "gofan"
)

func applyLevel(level string) error {
	// echo level 0 | tee /proc/acpi/ibm/fan

	stdin := &bytes.Buffer{}

	tee := exec.Command("tee", dev)
	tee.Stdin = stdin
	tee.Stderr = os.Stderr
	tee.Stdout = log.Writer()

	stdin.WriteString("level " + level + "\n")

	return tee.Run()
}

func checkIfShouldDowngrade(transitionWhenBelow float32, timeout time.Duration) <-chan error {
	c := make(chan error)
	ticker := time.NewTicker(timeout) // should dowgrade asap

	go func(transitionWhenBelow float32, callback chan<- error, ticker *time.Ticker) {
		for range ticker.C {
			temp, err := getTemp()
			if err != nil {
				callback <- err
				return
			}

			if temp > transitionWhenBelow {
				continue
			}

			callback <- nil
		}
	}(transitionWhenBelow, c, ticker)

	return c
}

func checkIfShouldUpgrade(ruleset *Ruleset, currentModeName string, timeout time.Duration) <-chan error {
	c := make(chan error)
	ticker := time.NewTicker(timeout)

	go func(ruleset *Ruleset, currentModeName string, callback chan<- error, ticker *time.Ticker) {
		for range ticker.C {
			temp, err := getTemp()
			if err != nil {
				c <- err
				return
			}

			mode, err := ruleset.findAppropriateMode(temp)
			notreached(err)

			if mode.Name == currentModeName {
				continue
			}

			c <- nil
		}
	}(ruleset, currentModeName, c, ticker)

	return c
}

// sometimes my os decides it's time for lift-off; even when the current cpu temperatur (and Elon Musk) says it's not. This occurs often when AC is plugged in
// in this case the current level as given by reading from the fan device does not equal the level of the mode we're currently in. In this case we should trigger a recheck on the currently best suited mode
func checkForUnmonitoredDeviceChanges(wantedCurrentLevel string) <-chan error {
	c := make(chan error)

	timout := time.Second * 10

	ticker := time.NewTicker(timout)

	go func(wantedCurrentLevel string, callback chan<- error, ticker *time.Ticker) {
		for range ticker.C {

			level, err := GetCurrentFanLevel(dev)
			if err != nil {
				c <- err
				return
			}

			if level == wantedCurrentLevel {
				continue
			}

			c <- nil
		}
	}(wantedCurrentLevel, c, ticker)

	return c
}

func logic(ruleset *Ruleset) error {

	var mode *Mode

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

		select {
		case err = <-checkIfShouldDowngrade(float32(mode.TransitionWhenBelow), ruleset.Timeouts.Downgrade.Duration):
			log.Printf("should downgrade\n")
		case err = <-checkIfShouldUpgrade(ruleset, mode.Name, ruleset.Timeouts.Upgrade.Duration):
			log.Printf("should upgrade\n")
		case err = <-checkForUnmonitoredDeviceChanges(mode.Level):
			log.Printf("umonitored device change\n")
		case <-time.After(ruleset.Timeouts.Standard.Duration):
			log.Printf("timed out after %v\n", ruleset.Timeouts.Standard.Duration)
		}

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
