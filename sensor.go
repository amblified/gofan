package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

type SensorsOutput struct {
	ThinkPadIsa `json:"thinkpad-isa-0000"`
	CoreTempIsa `json:"coretemp-isa-0000"`
}

type ThinkPadIsa struct {
	Temp1 struct {
		Input float64 `json:"temp1_input"`
	} `json:"temp1"`
	Temp2 struct {
		Input float64 `json:"temp2_input"`
	} `json:"temp2"`
	Temp3 struct {
		Input float64 `json:"temp3_input"`
	} `json:"temp3"`
	Temp4 struct {
		Input float64 `json:"temp4_input"`
	} `json:"temp4"`
	Temp5 struct {
		Input float64 `json:"temp5_input"`
	} `json:"temp5"`
	Temp6 struct {
		Input float64 `json:"temp6_input"`
	} `json:"temp6"`
	Temp7 struct {
		Input float64 `json:"temp7_input"`
	} `json:"temp7"`
	Temp8 struct {
		Input float64 `json:"temp8_input"`
	} `json:"temp8"`
}

type CoreTempIsa struct {
	Adapter string `json:"Adapter"`
	Core0   struct {
		Temp2Input     float32 `json:"temp2_input"`
		Temp2Max       float32 `json:"temp2_max"`
		Temp2Crit      float32 `json:"temp2_crit"`
		Temp2CritAlarm float32 `json:"temp2_crit_alarm"`
	} `json:"Core 0"`

	Core1 struct {
		Temp3Input     float32 `json:"temp3_input"`
		Temp3Max       float32 `json:"temp3_max"`
		Temp3Crit      float32 `json:"temp3_crit"`
		Temp3CritAlarm float32 `json:"temp3_crit_alarm"`
	} `json:"Core 1"`

	Core2 struct {
		Temp4Input     float32 `json:"temp3_input"`
		Temp4Max       float32 `json:"temp3_max"`
		Temp4Crit      float32 `json:"temp3_crit"`
		Temp4CritAlarm float32 `json:"temp3_crit_alarm"`
	} `json:"Core 2"`

	Core3 struct {
		Temp5Input     float32 `json:"temp5_input"`
		Temp5Max       float32 `json:"temp5_max"`
		Temp5Crit      float32 `json:"temp5_crit"`
		Temp5CritAlarm float32 `json:"temp5_crit_alarm"`
	} `json:"Core 3"`
}

func getInfo() (*SensorsOutput, error) {
	stdout := &bytes.Buffer{}

	sensors := exec.Command("sensors", "-j")
	sensors.Stdout = stdout
	sensors.Stderr = nil
	err := sensors.Run()
	if err != nil {
		return nil, err
	}

	var sensorsData SensorsOutput

	bytes := stdout.Bytes()

	err = json.Unmarshal(bytes, &sensorsData)
	if err != nil {
		return nil, err
	}

	return &sensorsData, nil
}

func getTemp() (float32, error) {
	info, err := getInfo()
	if err != nil {
		return -1, err
	}

	values := []float32{
		info.CoreTempIsa.Core0.Temp2Input,
		info.CoreTempIsa.Core1.Temp3Input,
		info.CoreTempIsa.Core2.Temp4Input,
		info.CoreTempIsa.Core3.Temp5Input,
	}

	current := values[0]
	for _, v := range values[1:] {
		if v > current {
			current = v
		}
	}

	return current, nil

	// info.CoreTempIsa.Core0.Temp2Input += info.CoreTempIsa.Core1.Temp3Input
	// info.CoreTempIsa.Core0.Temp2Input += info.CoreTempIsa.Core2.Temp4Input
	// info.CoreTempIsa.Core0.Temp2Input += info.CoreTempIsa.Core3.Temp5Input

	// return info.CoreTempIsa.Core0.Temp2Input / 4, nil
}
