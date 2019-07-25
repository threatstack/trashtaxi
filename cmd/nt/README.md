# nt
"Sudo? More like su_dont_!" A member of the 
[Trash Taxi](https://github.com/threatstack/trashtaxi) group of commands.

Reducing the amount of "You Only Live Once" (YOLO) actions in your
infrastructure is challenging, because sometimes you need to log into a server
to run odd commands. How can you enable engineers to safely use a root shell
and, in time, ensure that the machine is definitely restored to a known good
state?

`nt` is a wrapper around bash that you can add to your sudo configuration.
Before it spawns a root shell, it will send an HTTP request containing the
instance's AWS Identity Document to the `garbaged` server, which marks the host
for later termination.

## Building, installing, pre-requisites

You'll need to be running in AWS to get the most out of this. It could be used
on other providers, but you'd need to figure out the "remediation" action.

You can use `go build` and drop the output wherever you wish. Then, add it to
your sudo configuration. You'll need to drop a configuration file -
`/etc/nt.json` is where it will expect to find it. The file takes two variables,
the `garbaged` endpoint and the preferred shell to spawn.

```
{
  "endpoint": "https://taxi.tls.zone"
  "shell": "/bin/bash"
}
```

At this point, you're ready to run `sudo nt` and get a shell. If `garbaged` is
set up correctly, `nt` will make a request and the host will be marked for
cleanup.

## Contributing

Contributions and modifications welcome.
