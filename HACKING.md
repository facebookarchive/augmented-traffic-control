# Hacking on ATC

## Building the development environment

ATC uses netlink under the hood to shape traffic, as such, ATC only runs on Linux.

ATC uses golang, glide, and sqlite3. You must have a relatively recent version of go (1.7, 1.8) as well
as glide (for dependendency management) and the sqlite3 library and development files.

To make it easier to deploy a development environment, we leverage [docker](https://www.docker.com/) and/or optionally [docker-compose](https://docs.docker.com/compose/)

Installing and configuring docker is out of the scope of this document, but plenty of documentation for your favorite operating system can be found online at https://docs.docker.com/installation/

### Using docker

The `Dockerfile` provided in this repo is pretty limited and basically build from golang:latest image. To build the ATC image, you need to run:

```
make docker-build
```

Once you have the image build, you can get a shell in the container by running:

```
make docker-run
```

This will give you a bash prompt in the docker container and will set you in the working directory: `/gopath/src/github.com/facebook/augmented-traffic-control`.
The root of the atc project will be mounted within the container. From there, any changes that you do in ATC repo, will be readily available in the container.
Also, port 8080 and 9090 are mapped from the host to the container allowing you to access both the atc daemon and the atc ui through the host.

### Using docker-compose

[docker-compose](https://docs.docker.com/compose/) is a tool that builds on top of docker to simplify container deployments and configuration. A `docker-compose.yml` file is provided allowing to get a prompt in a container by running the following:

```
docker-compose run --service-ports atc bash
```

## Building ATC

Once you have your development environment setup, you can build/install atcd by running:

```
glide i
make
```

This will download all atcd dependencies as well as build it, run tests and install it.

### Making a static build

Golang binary can easily be built sttatically, allowing to copy the binary around without even have to worry about libc. To build ATC statically, set the `STATICBUILD` environment variable to somehting as in:

```
STATICBUILD=1 make
```

## Starting ATC

After you have built ATC, you can run it by running (from within the default workdir):

```
./bin/atcd &
./bin/atc_api
```

You can now access the ATC UI from http://<dockerip>:8080/

## Building the UI

A pre-built .js is provided in `static/js`, but if you want to modify it, you will need to change the [JSX](https://facebook.github.io/react/docs/getting-started.html) files in `src/react/jsx` and re-generate `static/js/index.js`.

There is a bunch of tools that need to be installed in order to be able to generate this `index.js` that is consumable by a web browser. At a high level, we need `npm`, `babel`, `react` and `browserify`.

The steps below are going to make use of [docker-compose](https://docs.docker.com/compose/) to set up the dev environment. If you do not want to use docker, the instructions will still apply but for the docker part (e.g you dont need the `docker-compose run node` bit).

To build the docker image used to develop on jsx run:
```
docker-compose build node
```

Once you have the image built, you can install the dependencies:
```
docker-compose run node make npm_env
```

To build the js file:
```
docker-compose run node make static/js/index.js
```

If you want to have the `static/js/index.js` file automatically generated as you modify the jsx files, you can run:
```
docker-compose run node bash -c 'cd src/react; npm run watch'
```


To run the UI linter:
```
docker-compose run node make lint-ui
```
