# sfv

[![Build Status](https://travis-ci.org/mpolden/sfv.svg)](https://travis-ci.org/mpolden/sfv)

sfv is a [Go](http://golang.org) package for verifying
[SFV files](https://en.wikipedia.org/wiki/Simple_file_verification).

## Installation

`$ go get github.com/mpolden/sfv`

## Example

```go
package main

import (
	"github.com/mpolden/sfv"
	"log"
)

func main() {
	sfv, err := sfv.Read("/path/to/file.sfv")
	if err != nil {
		log.Fatal(err)
	}

	ok, err := sfv.Verify()
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		log.Print("All files are OK!")
	}
	for _, c := range sfv.Checksums {
		log.Printf("%+v", c)
	}
}
```
