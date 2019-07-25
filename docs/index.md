Trash Taxi is a lifecycle management tool that helps reduce configuration
drift in your environment, by terminating servers when arbitrary manual
commands have been executed on them.

## Why would I need this?
We know that manual changes to systems can introduce security and availability 
issues. But an organization's observability tooling doesn't exist at day zero; 
it grows over time. Often, the tooling gets built as a result of learning how a system
responds to issues under load, in production. 

Trash Taxi is a way to balance an organization's need to allow some amount 
of unrestricted access to a machine, while ensuring that the machine is
terminated at a later time. Developers get the information they need,
Operations can share responsibility, and Security can sleep (slightly more)
soundly at night.

## How's it work?
When a user needs a shell, they run `sudo nt` (_"sudo? more like su 
dont!"_ - we're punny here). A prompt asks the engineer if they're sure 
they want to mark the host for later deletion. Then they get their shell.

`nt` sends the Amazon Instance Identity Document to `garbaged`, which handles
termination. You can schedule _trash pickups_ (terminations) using a
variety of methods (cloudwatch events, AWS IoT button, API call). 

## Wow, that sounds scary!
Trash Taxi ships with a few safety features built in:
* You can schedule _trash holidays_ based on a EC2 `Role` or `Type` tag, allowing
  you to track when folks use `nt` on a stateful host without terminating it.
* Trash Taxi will only remove one of each role on every run, to ensure you
  don't take out an entire role with one misplaced command.
* You can configure how many hosts you want to have terminated in any given
  run, and a forced waiting period between runs as well.

## Hop in...
Interested in giving Trash Taxi a try?

* Take a look at the [Getting Started](getstarted.html) guide.
* Make your [Configuration](config.html) file.
* Utilize the `garbaged` [API](api.html) to pick up the trash!
