watch
=====

watches the current directory for changes, then executes a command

install
----------

Assuming go is installed:

`go get github.com/jakecoffman/watch`

running
----------

`cd` into your desired directory and run: `$GOPATH/bin/watch`

or if `$GOPATH/bin` is in your `$PATH`, simply: `watch`

By default, watch will run `go test`. To run something else, give it arguments: `watch go build`
