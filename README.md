# Trash Taxi
![The Trash Taxi logo is of a sloth driving a taxi cab with servers on fire in the trunk](docs/taxicircle.png)

Trash Taxi is a lifecycle management tool that helps reduce configuration drift in your environment, by terminating servers when arbitrary manual commands have been executed on them.

Most documentation is available at [https://trash.taxi](https://trash.taxi).

## Building and installing
You can use `go build` and drop the output wherever you wish. `garbaged`'s configuration file (by default) lives in `/etc/garbaged.json` but you can reassign it using the `GARBAGED_CONFIG` environment variable.

The `nt` command is available in `cmd/nt`, along with information on how to configure `nt` as well.

There is a command line API utility named `tt` which you'll find in `cmd/tt`. 

## Issues, Questions, Comments, etc. 
Trash Taxi is not supported as part of a F5 service contract, and is maintained as we have time. Your best route to getting questions answered about Trash Taxi is by opening a GitHub issue.

## Contributing
Before you start contributing to any project sponsored by F5, Inc. (F5) on GitHub, you will need to sign a Contributor License Agreement (CLA). This document can be provided to you once you submit a GitHub issue that you contemplate contributing code to, or after you issue a pull request.

If you are signing as an individual, we recommend that you talk to your employer (if applicable) before signing the CLA since some employment agreements may have restrictions on your contributions to other projects. Otherwise by submitting a CLA you represent that you are legally entitled to grant the licenses recited therein.

If your employer has rights to intellectual property that you create, such as your contributions, you represent that you have received permission to make contributions on behalf of that employer, that your employer has waived such rights for your contributions, or that your employer has executed a separate CLA with F5.

If you are signing on behalf of a company, you represent that you are legally entitled to grant the license recited therein. You represent further that each employee of the entity that submits contributions is authorized to submit such contributions on behalf of the entity pursuant to the CLA.