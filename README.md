# Augmented Traffic Control

[![GoDoc](https://godoc.org/github.com/facebook/augmented-traffic-control?status.svg)](https://godoc.org/github.com/facebook/augmented-traffic-control)
[![build-status-image]][travis]

## WARNING

This version of ATC is experimental and not yet officially released. We welcome
feedback on this version of ATC as we continue working on it, but you should
not use this version of ATC unless you are prepared to deal with bugs or
feature changes.


Full documentation for the project is available at
[http://facebook.github.io/augmented-traffic-control/](http://facebook.github.io/augmented-traffic-control/).


## Overview

Augmented Traffic Control (ATC) is a tool to simulate network conditions. It allows controlling the connection that a device has to the internet.
Developers can use `ATC` to test their application across varying network conditions, easily emulating mobile, and severely impaired networks.

Aspects of the connection that can be controlled include:

* bandwidth
* latency
* packet loss
* corrupted packet rates
* packet ordering

In order to be able to shape the network traffic, ATC must be running on a device that routes the traffic and sees the real IP address of the device,
for example, your network gateway. This also allows any devices that route through `ATC` to be able to shape their traffic.
Traffic can be shaped/unshaped using a web interface allowing any devices with a web browser to use `ATC` without the need for a client application.

ATC is made of two components that interact together:
* `atcd`: A low-level thrift interface which is responsible for directly setting/unsetting traffic shaping.
* `atc_api`: A RESTful HTTP interface to `atcd`. Proxies requests to an `atcd` thrift server.

![ATC architecture][atc_architecture]


## Installing ATC

The fact that `ATC` is split in multiple packages allows for multiple deployment scenarios.
However, deploying all the packages on the same host is the simplest way to setup ATC.

### Requirements

ATC requires a  working [Golang](https://golang.org/) toolchain.

Most linux distributions have binary packages available for go.

For example:

### Installing from Source

First, setup a GOPATH. If you aren't doing normal golang development, you can run
`setup.sh` within the ATC project and a GOPATH will be setup locally specifically for ATC.
After running the script, you must `export GOPATH="$(pwd)/.gopath"` before compiling ATC.

Clone the ATC code somewhere on your machine and run `make` within the project:

``` bash
git clone git@github.com:facebook/augmented-traffic-control.git
cd augmented-traffic-control
./setup.sh
export GOPATH="$(pwd)/.gopath"
make
```

This will download all of ATC's dependencies and compile `atcd` and `atc_api` into `bin/`.

Note that downloading the dependencies may take a while depending on your connection speed.

You can then run `sudo make install` to copy these binaries into `/usr/local/bin/`, but this isn't required to run ATC.

MARKER

## ATC Code Structure

ATC source code is available under the [atc](atc/) directory, it is currently composed of:

* [atc_thrift](atc/atc_thrift) the thrift interface's library
* [atcd](atc/atcd) the ATC daemon that runs on the router doing the traffic shaping
* [django-atc-api](atc/django-atc-api) A django app that provides a RESTful interface to `atcd`
* [django-atc-demo-ui](atc/django-atc-demo-ui) A django app that provides a simple demo UI leveraging the RESTful API
* [django-atc-profile-storage](atc/django-atc-profile-storage) A django app that allows saving shaping profiles to DB allowing users to select their favorite profile from a list instead of re-entering all the profile details every time.


The [chef](chef/) directory contains 2 chef cookbooks:

* [atc](chef/atc/) A cookbook to deploy ATC. It also allows to deploy ATC in a Virtual Box VM in order to develop on ATC.
* [atclient](chef/atcclient) Set up a Linux Desktop VM that can be used to test shaping end to end.

### atcd

`atcd` is the daemon that runs on the router that does the shaping. Interaction with the daemon is done using [thrift](https://thrift.apache.org/). The interface definition can be found in [atc_thrift.thrift](atc/atc_thrift/atc_thrift.thrift).

### atc_thrift

`atc_thrift` defines the thrift interface to communicate with the `atcd` daemon.

### django-atc-api

`django-atc-api` is a django app that provide a REST API to the `atcd` daemon. Web applications, command line tools can use the API in order to shape/unshape traffic.

### django-atc-demo-ui

`django-atc-demo-ui` is a simple Web UI to enable/disable traffic shaping. The UI is mostly written in [React](http://facebook.github.io/react/)

### django-atc-profile-storage

`django-atc-profile-storage` allows saving profiles to DB. A typical use case will be to save a list of predefined/often used shaping settings that you want to be able to accessing in just a few clicks/taps.

## Developing on ATC

To make ATC development easier, we use Virtual Box and Vagrant to provision and run a VM that will run the ATC daemon and the ATC UI from your git checkout.

Interacting with ATC will only shape the traffic within the VM and not on the host.

### Setting up the environment

Note: vagrant is an easy way to set up a test environment, but virtualization will produce different results than a setup on bare-metal. We recommend using vagrant only for testing/development and using bare-metal for setups which require realistic shaping settings.

You will need to install VirtualBox, Vagrant and a couple of plugins:

* [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
* [Vagrant](https://www.vagrantup.com/downloads.html)
* [Chef DK](https://downloads.chef.io/chef-dk/)
* Install some vagrant plugins:
 * vagrant plugin install vagrant-berkshelf --plugin-version '>= 2.0.1'
 * vagrant plugin install vagrant-omnibus
* Clone this repo: git clone git@github.com:facebook/augmented-traffic-control.git atc

### Running ATC

Once in the repo, go to the `chef/atc` directory and run:

``` bash
vagrant up trusty
```

This will take some time before it completes, once the VM is provision, SSH into it:

``` bash
vagrant ssh trusty
```

You should now be able to access ATC at: http://localhost:8080/

### Using the Sample Profiles

Once you've got ATC up and running, you can run the script `utils/restore-profiles.sh` to setup the set of default profiles.

The script needs to be passed a `hostname:port` with the location of your ATC instance:

    utils/restore-profiles.sh localhost:8080

After doing this, you should see the 10 sample profiles listed below in your ATC instance:

- `2G - Developing Rural`
- `2G - Developing Urban`
- `3G - Average`
- `3G - Good`
- `Cable`
- `DSL`
- `Edge - Average`
- `Edge - Good`
- `Edge - Lossy`
- `No Connectivity`

Naturally, you cannot improve your natural network speed by selecting a faster profile than your service. For example, selecting the `Cable` profile will not make your network faster if your natural connection speed resembles DSL more closely.

### Hacking on the code

Hacking on ATC is done from the host and tested in the VM. In order to reflect the changes, you will need to start the services manually.

Both `atcd` and `atcui` have their python libraries installed in a *python virtualenv* so you will need to activate the environment in order to be able to run the services.

The *virtualenv* is installed in */usr/local/atc/venv/bin/activate* .

``` bash
source /usr/local/atc/venv/bin/activate
```

#### Running the daemon

The `atcd` daemon is running under the root user privileges, all operations below needs to be done as root.

To run the daemon manually, first make sure it is not running in the background:

``` bash
service atcd stop
```

And run the daemon:

``` bash
atcd
```

Once you are happy with your changes and you want to test them, you will need to kill the daemon and restart it in order to apply the changes.

#### Running the API/UI

This is a django project and, when running the django built-in HTTP server, will detect code changes and reload automatically.

To run the HTTP REST API and UI:

``` bash
cd /var/django && python manage.py runserver 0.0.0.0:8000
```

[atc_architecture]: https://facebook.github.io/augmented-traffic-control/images/atc_overview.png
[build-status-image]: https://travis-ci.org/facebook/augmented-traffic-control.svg?branch=master
[travis]: https://travis-ci.org/facebook/augmented-traffic-control?branch=master
[pypi-version]: https://pypip.in/version/atcd/badge.svg
[pypi]: https://pypi.python.org/pypi/atcd

