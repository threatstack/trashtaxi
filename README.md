# Trash Taxi
![The Trash Taxi logo is of a sloth driving a taxi cab with servers on fire in the trunk](docs/taxicircle.png)

Trash Taxi is a lifecycle management tool that helps reduce configuration
drift in your environment, by terminating servers when arbitrary manual
commands have been executed on them.

Most documentation is available at [https://trash.taxi](https://trash.taxi).

## Building and installing
You can use `go build` and drop the output wherever you wish. `garbaged`'s
configuration file (by default) lives in `/etc/garbaged.json` but you can
reassign it using the `GARBAGED_CONFIG` environment variable.

The `nt` command is available in `cmd/nt`, along with information on how
to configure `nt` as well.

There is a command line API utility named `tt` which you'll find in `cmd/tt`. 

## Issues, Questions, Comments, etc. 
Trash Taxi is not supported as part of a Threat Stack service contract, 
and is maintained as we have time. Your best route to getting questions answered
about Trash Taxi is by opening a GitHub issue.

## Contributing
We welcome your contribution to Trash Taxi! Fortunately, this project is small: 
You'll find the bulk of the server code and endpoint handlers under `server/`. 
If you need to add configuration flags or variables you'll find configuration structs
such under `config/`. If you're interested in finding out if some feature
work would be helpful, please open an issue.
