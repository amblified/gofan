package main

import (
	"bytes"
	"log"
	"os/exec"
	"time"
)

func applyLevel(level string) error {
	// echo level 0 | tee /proc/acpi/ibm/fan

	stdin := &bytes.Buffer{}

	dev := "/proc/acpi/ibm/fan"
	tee := exec.Command("tee", dev)
	tee.Stdin = stdin
	tee.Stderr = log.Writer()
	tee.Stdout = log.Writer()

	stdin.WriteString("level " + level + "\n")

	return tee.Run()
}

func checkIfShouldDowngrade(transitionWhenBelow float32) <-chan error {
	c := make(chan error)
	ticker := time.NewTicker(time.Second * 6) // should dowgrade asap

	go func(transitionWhenBelow float32, callback chan<- error, ticker *time.Ticker) {
		for range ticker.C {
			temp, err := getTemp()
			if err != nil {
				callback <- err
			}

			if temp > transitionWhenBelow {
				continue
			}

			callback <- nil
		}
	}(transitionWhenBelow, c, ticker)

	return c
}

func checkIfShouldUpgrade(ruleset *Ruleset, currentModeName string) <-chan error {
	c := make(chan error)
	ticker := time.NewTicker(time.Second * 24)

	go func(ruleset *Ruleset, currentModeName string, ticker *time.Ticker) {
		for range ticker.C {
			temp, err := getTemp()
			if err != nil {
				c <- err
			}

			mode, err := ruleset.findAppropriateMode(temp)
			notreached(err)

			if mode.Name == currentModeName {
				continue
			}

			c <- nil
		}
	}(ruleset, currentModeName, ticker)

	return c
}

// sometimes my os decides it's time for lift-off; even when the current cpu temperatur says it's not. This occurs often when AC is plugged in
func fanToFastForCurrentTemp() <-chan error {
	c := make(chan error)
	return c
}

func logic(ruleset *Ruleset) error {

	var mode *Mode

	for {
		log.Printf("\n")
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

		timeout := time.Minute * 1

		select {
		case err = <-checkIfShouldDowngrade(float32(mode.TransitionWhenBelow)):
			log.Printf("should downgrade\n")
		case err = <-checkIfShouldUpgrade(ruleset, mode.Name):
			log.Printf("should upgrade\n")
		case err = <-fanToFastForCurrentTemp():
		case <-time.After(timeout):
			log.Printf("timed out after %v\n", timeout)
		}

		if err != nil {
			return err
		}
	}
}

func main() {
	if err := logic(DefaultRuleset); err != nil {
		log.Fatal(err)
	}
}
