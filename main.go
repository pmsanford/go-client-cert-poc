package main

import (
	"flag"
	"fmt"
)


func main() {
	serverFlag := flag.Bool("s", false, "Run server")
	clientFlag := flag.Bool("c", false, "Run client")
	genFlag := flag.Bool("g", false, "Generate cert")
	testFlag := flag.Bool("t", false, "Test")

	flag.Parse()

	if *serverFlag {
		runserver()
	} else if *clientFlag {
		runclient()
	} else if *genFlag {
		makecert()
	} else if *testFlag {
		testCA()
	} else {
		fmt.Println("Try -s, -c, -g, or -h")
	}
}