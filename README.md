# gofan
A small (customizeable) fan control for linux.

## Dependencies
There are no system dependencies.
Library dependencies are listed in [go.mod](https://gitlab.com/malte-L/go-fan/-/blob/master/go.mod) (currently only `golang.org/x/sync`). 


## Example Usage

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
Currently the client is written specifically for my needs. Even tho you can write your own client using the functionality of [/fan](https://gitlab.com/malte-L/go-fan/-/tree/master/fan) it should be easier in the future to customize the behaviour of the fan controller.
