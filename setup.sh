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
    # github.com/vishvananda/netns and github.com/alecthomas/assert
    # are used for unittests.
    for x in $(depends) \
            "github.com/vishvananda/netns" \
            "github.com/alecthomas/assert"
    do
        echo "$x"
        go get "$x"
    done
}

depends() {
    # Special since it's a build-time dependency and isn't imported by any code
    echo "github.com/jteeuwen/go-bindata/go-bindata"
    go list -f '{{range .Imports}}{{.}}{{"\n"}}{{end}}' "$SRC/daemon" "$SRC/atcd" "$SRC/atc" "$SRC/api" "$SRC/atc_api" "$SRC/shaping" | sort -u | fgrep . | grep -v "augmented-traffic-control"
}

if [ ! "$CI" == true ]; then
    GOPATH="$(pwd)/.gopath/"
    export GOPATH
    make_gopath
fi
get_depends
