#!/usr/bin/env bash

PROJECT=github.com/facebook/augmented-traffic-control
SRC="$PROJECT/src"

make_gopath() {
    echo "Setting up GOPATH in $GOPATH"
    mkdir -p "$GOPATH/src/$PROJECT"
    ln -s "$(pwd)/src" "$GOPATH/src/$PROJECT"
}

get_depends() {
    echo "Fetching dependencies. (This might take a few minutes)."
    for x in $(depends) ; do
        echo "$x"
        go get "$x"
    done
    # Used to run unittests
	x="github.com/vishvananda/netns"
	echo "$x"
	go get "$x"
}

depends() {
    # Special since it's a build-time dependency and isn't imported by any code
    echo "github.com/jteeuwen/go-bindata/go-bindata"
    go list -f '{{range .Imports}}{{.}}{{"\n"}}{{end}}' "$SRC/daemon" "$SRC/atcd" "$SRC/api" "$SRC/atc_api" "$SRC/shaping" | sort -u | fgrep . | grep -v "augmented-traffic-control"
}

if [ ! "$CI" == true ]; then
    GOPATH="$(pwd)/.gopath/"
    export GOPATH
    make_gopath
fi
get_depends
