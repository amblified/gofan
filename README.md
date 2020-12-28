# gofan
A small (customizeable) fan control for linux.

## Dependencies
The client uses [lm_sensors](https://hwmon.wiki.kernel.org/lm_sensors) for fetching cpu temperature.
Library dependencies are listed in [go.mod](https://github.com/amblified/gofan/blob/master/go.mod) (currently only `golang.org/x/sync`). 

## Configuration
The client can be configured using a ruleset. The ruleset is a json file containing timeouts and modes. The timeouts are used for detecting certain events (see next section) at a given frequency. 
```json
"timeouts": {
  "standard": "1m",
  "upgrade": "20s",
  "downgrade": "6s",
  "unmonitored_change": "10s"
},

```
An example mode could look like this: 
```json
{
  "name": "light",
  "starting_at": 55,
  "transition_when_below": 50,
  "level": "1"
},
```
`name` is simply an identifier. `starting_at` specifies a temperature (in °C) which has to be surpassed in order for the mode to be applied. When the temperature falls below `transition_when_below` the client will reevaluate the best fitting mode. `level` is the fan level which is applied when the prgram enters this mode. For my needs, the server currently only accepts 1-6 and "auto", but theoretically this could be any level which the fan device accepts (`cat <DEVICE>`). For example `cat /proc/acpi/ibm/fan` yields 
```
status:		disabled
speed:		0
level:		0
commands:	level <level> (<level> is 0-7, auto, disengaged, full-speed)
commands:	enable, disable
commands:	watchdog <timeout> (<timeout> is 0 (off), 1-120 (seconds))
```

## Example Usage


The client will look for a configuration file in the config directory of your system (usually `.config`), it assumes the full path to be `<path to config-dir>/gofan/rules`. The path to a ruleset can be overwritten with the `-rules` flag.

You can use the [example_rules.json](https://github.com/amblified/gofan/blob/master/examples/example_rules.json) if you want to get started.

Start the server with:
`sudo ./gofan_server -stream=gofan.sock -dev=/proc/acpi/ibm/fan`

and the client with:
`./gofan_client -stream=gofan.sock -dev=/proc/acpi/ibm/fan`.


The client will now detect different events:
- **ShouldUpgrade**: there exists a better suited (a faster) fan mode for the current temperature
- **ShouldDowngrade**: there exists a better suited (a slower) fan mode for the current temperature
- **UnmonitoredDeviceChange**: another party changed the current fan mode. This could be achieved using for example `echo level 2 | sudo tee /proc/acpi/ibm/fan`. In my case my laptop sometimes turns on the fan on max speed upon plugging in AC.
- **TimedOut**: using the example config file, the client times out after 1 minute (`"standard": "1m"`).

If any of the listed events fires, the client will determine the best suited fan mode for the current cpu-temperature and send it to the server.
The server validates the received fan level and forwards it into the given `-dev` file (in example above: `/proc/acpi/ibm/fan`).

## Future Plans
Currently the client is written specifically for my needs. Even though you can write your own client using the functionality of [/fan](https://github.com/amblified/gofan/tree/master/fan) it should be easier in the future to customize the behaviour of the fan controller.
