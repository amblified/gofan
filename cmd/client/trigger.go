package main

/* listed in this file are functions which may give the main go routine a
** hint that there might be a better suited mode available.
** This triggering is done via the callback channel. nil indicating a trigger,
** an error indicating something went wrong.
** As of for now, the main go routine aborts upon error */

import (
	"log"
	"time"

	"gitlab.com/malte-L/go_fan/fan"
)

func checkIfShouldDowngrade(currentMode **fan.Mode, ticker *time.Ticker, callback chan<- error) {
	for range ticker.C {
		if *currentMode == nil {
			continue
		}

		temp, err := fan.GetTemp()
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

func notreached(err error) {
	if err != nil {
		_notreached()
	}
}

func _notreached() {
	panic("this code should not be reached")
}

func checkIfShouldUpgrade(currentMode **fan.Mode, ruleset *fan.Ruleset, ticker *time.Ticker, callback chan<- error) {
	for range ticker.C {
		if *currentMode == nil {
			continue
		}

		temp, err := fan.GetTemp()
		if err != nil {
			callback <- err
			return
		}

		mode, err := ruleset.FindAppropriateMode(temp)
		notreached(err)

		if mode.Name == (*currentMode).Name {
			continue
		}

		if mode.StartingAt < (*currentMode).StartingAt {
			continue
		}

		log.Printf("should upgrade (checked with temperature: %f°C)\n", temp)
		callback <- nil
	}
}

// sometimes my os decides it's time for lift-off; even when the current cpu temperatur (and Elon Musk) says it's not. This occurs often when AC is plugged in
// in this case the current level as given by reading from the fan device does not equal the level of the mode we're currently in. In this case we should trigger a recheck on the currently best suited mode
func checkForUnmonitoredDeviceChanges(currentMode **fan.Mode, ticker *time.Ticker, callback chan<- error) {
	for range ticker.C {
		if *currentMode == nil {
			continue
		}

		level, err := fan.GetCurrentFanLevel(*devPath)
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

func checkForTimeout(ruleset fan.Ruleset, ticker *time.Ticker, callback chan<- error) {
	for range ticker.C {
		log.Printf("timed out after %v\n", ruleset.Timeouts.Standard.Duration)
		callback <- nil
	}
}
