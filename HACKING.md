# Hacking on ATC

## Building the development environment

ATC uses netlink under the hood to shape traffic, as such, ATC only runs on Linux.

To make it easier to deploy a development environment, we leverage [docker](https://www.docker.com/) and/or optionally [docker-compose](https://docs.docker.com/compose/)

Installing and configuring docker is out of the scope of this document, but plenty of documentation for your favorite operating system can be found online at https://docs.docker.com/installation/

### Using docker

The `Dockerfile` provided in this repo is pretty limited and basically build from golang-1.5.1 image. To build the ATC image, you need to run:

```
docker build -t atc .
```

Once you have the image build, you can get a shell in the container by running:

```
docker run -ti --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp -p 9090:9090 -p 8080:8080 atc bash
```

This will give you a bash prompt in the docker container and will set you in the working directory: `/usr/src/myapp`. The root of the atc project will be mounted within the container. From there, any chnages that you do in ATC repo, will be readily available in the container.
Also, port 8080 and 9090 are mapped from the host to the container allowing you to access both the atc daemon and the atc ui through the host.

### Using docker-compose

[docker-compose](https://docs.docker.com/compose/) is a tool that builds on top of docker to simplify container deployments and configuration. A `docker-compose.yml` file is provided allowing to get a prompt in a container by running the following:

```
docker-compose run --service-ports atc bash
```

## Building ATC

Once you have your development environment setup, you can build/install atcd by running:

```
make
```

This will download all atcd dependencies as well as build it, run tests and install it.

## Starting ATC

After you have built ATC, you can run it by running (from within the default workdir):

```
./bin/atcd &
./bin/atc_api
```

You can now access the ATC UI from http://<dockerip>:8080/
