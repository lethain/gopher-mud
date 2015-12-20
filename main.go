package main

import (
	"github.com/lethain/gopher-mud/mud"
	"flag"
)

var loc = flag.String("loc", ":9000", "location:port to run server")

func main() {
	flag.Parse()
	ms := mud.MudServer{Loc: *loc}
	ms.ListenAndServe()
}
