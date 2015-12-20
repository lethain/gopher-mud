package main

import (
	"flag"
	"github.com/lethain/gopher-mud/mud"
)

var loc = flag.String("loc", ":9000", "location:port to run server")

func main() {
	flag.Parse()
	ms := mud.MudServer{Loc: *loc}
	ms.ListenAndServe()
}
