//go:build !unix

package main

import "errors"

var errDetachUnsupported = errors.New("detaching is not supported on this platform")

func detach(root string) error { return errDetachUnsupported }
