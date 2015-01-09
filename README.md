# Air Traffic Control

Air Traffic Control (ATC) is a tool to allow controlling the connection that a device has to the internet. Aspects of the connection that can be controlled include:

* bandwidth
* latency
* packet loss
* corrupted packets
* packets ordering

## Installing ATC

`TODO / Improve`

ATC components can be installed via `pip`.

To install the ATC daemon `atcd`:

<code>pip install atcd</code>


Both the `django-atc-api` and `django-atc-demo-ui` requires you to have a [Django](https://www.djangoproject.com/) project set up. Once you have the django project set, you can install the modules and enable them by adding them to the `INSTALLED_APPS` in your project' `settings.py`.

To install the REST API app:

<code>pip install django-atc-api</code>

And add:

        'atc_api',
        'rest_framework',

to `INSTALLED_APPS` .


For the Demo UI app:

<code>pip install django-atc-demo-ui</code>

And add:

        'atc_demo_ui',


to `INSTALLED_APPS`.

You will also need to route the queries to those apps by adding something along the line of:

        url(r'^api/v1/', include('atc_api.urls')),
        url(r'^atc_demo_ui/', include('atc_demo_ui.urls')),

in your project's `urls.py`.

## ATC Code Structure

ATC source code is available under the [atc](atc/) directory, it is currently composed of:

* [atc_thrift](atc/atc_thrift) the thrift interface's library
* [atcd](atc/atcd) the ATC daemon that runs on the router doing the traffic shaping
* [django-atc-api](atc/django-atc-api) A django app that provides a RESTful interface to `atcd`
* [django-atc-demo-ui](atc/django-atc-demo-ui) A django app that provides a simple demo UI leveraging the RESTful API


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

`django-atc-demo-ui` is a simple Web UI to enable/disable traffic shaping.

## Developing on ATC

To make ATC development easier, we use Virtual Box and Vagrant to provision and run a VM that will run the ATC daemon and the ATC UI from your git checkout.

Interacting with ATC will only shape the traffic within the VM and not on the host.

### Setting up the environment

You will need to install VirtualBox, Vagrant and a couple of plugins:

* [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
* [Vagrant](https://www.vagrantup.com/downloads.html)
* [Chef DK](https://downloads.chef.io/chef-dk/)
* Install some vagrant plugins:
 * vagrant plugin install vagrant-berkshelf --plugin-version '>= 2.0.1'
 * vagrant plugin install vagrant-omnibus
* Clone this repo: git clone git@github.com:facebook/air-traffic-control.git atc

### Running ATC

Once in the repo, go to the `chef/atc` directory and run:

<code>vagrant up atccentos</code>

This will take some time before it completes, once the VM is provision, SSH into it:

<code>vagrant ssh atccentos</code>

And initialize the django application:

<code>sudo /usr/local/bin/atcui-setup</code>

You should now be able to access ATC at: http://localhost:8080/atc/

### Hacking on the code

Hacking on ATC is done from the host and tested in the VM. In order to reflect the changes, you will need to start the services manually.

Both `atcd` and `atcui` have their python libraries installed in a *python virtualenv* so you will need to activate the environment in order to be able to run the services.

The *virtualenv* is installed in */usr/local/atc/venv/bin/activate* .

<code>source /usr/local/atc/venv/bin/activate</code>


#### Running the daemon

The `atcd` daemon is running under the root user privileges, all operations below needs to be done as root.

To run the daemon manually, first make sure it is not running in the background:

<code>/etc/init.d/atcd stop</code>

And run the daemon:

<code>atcd</code>

Once you are happy with your changes and you want to test them, you will need to kill the daemon and restart it in order to apply the changes.

#### Running the API/UI

The `atc_api` and `atc ui` are currently run with root privileges. This is a django project and, when running the django built-in HTTP server, will detect code changes and reload automatically.

To run the HTTP REST API and UI:

<code>
cd /var/django && python manage.py runserver 0.0.0.0:8000
</code>

