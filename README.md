watch
=====

[![Build Status](https://secure.travis-ci.org/jakecoffman/watch.png?branch=master)](http://travis-ci.org/jakecoffman/watch)

Watches the current directory and subdirectories for changes, then executes a command.

Honors the .gitignore file so it won't fire when .git files change.

install
----------

Assuming go is installed:

`go get github.com/jakecoffman/watch`

running
----------

`cd` into your desired directory and run: `$GOPATH/bin/watch`

or if `$GOPATH/bin` is in your `$PATH`, simply: `watch`

By default, watch will run `go test`. To run something else, give it arguments: `watch go build`
