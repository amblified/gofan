package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	DefaultRuleset     *Ruleset
	DefaultRulesetPath = "./rules.json"
)

type Ruleset struct {
	Timeouts struct {
		Standard          Duration `json:"standard"`
		Upgrade           Duration `json:"upgrade"`
		Downgrade         Duration `json:"downgrade"`
		UnmonitoredChange Duration `json:"unmonitored_change"`
	} `json:"timeouts"`

	Modes []Mode `json:"modes"`
}

type Mode struct {
	Name                string `json:"name"`
	Level               string `json:"level"`
	StartingAt          int    `json:"starting_at"`
	TransitionWhenBelow int    `json:"transition_when_below"`
}

func notreached(err error) {
	if err != nil {
		_notreached()
	}
}

func _notreached() {
	panic("this code should not be reached")
}

func (m Mode) String() string {
	bytes, err := json.Marshal(m)
	notreached(err)

	return string(bytes) + "\n"
}

func readRuleset(path string) (*Ruleset, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rules Ruleset
	bytes, err := ioutil.ReadAll(file)
	json.Unmarshal(bytes, &rules)

	return &rules, nil
}

// go is fun, innit?
func init() {
	var err error

	DefaultRulesetPath, err = filepath.Abs(DefaultRulesetPath)
	notreached(err)
	log.Printf("loading rules from %q\n", DefaultRulesetPath)
	DefaultRuleset, err = readRuleset(DefaultRulesetPath)
	if err != nil {
		log.Printf("fatal error occured when trying to load ruleset. aborting..")
		os.Exit(1)
	}
}

// TODO: NOTE: that the default mode MUST (currently) be the first one listed in the json file; otherwise this method might return falsy results. FIX: maybe sort the modes?
func (r Ruleset) findAppropriateMode(temp float32) (*Mode, error) {
	if len(r.Modes) < 1 {
		_notreached()
	}

	mode := r.Modes[0]
	for _, m := range r.Modes[1:] {
		if float32(m.StartingAt) < temp && m.StartingAt > mode.StartingAt {
			mode = m
		}
	}
	return &mode, nil
}
