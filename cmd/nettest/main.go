package main

import (
	"flag"
	"log"
	"net"

	"github.com/DoubleDi/tcprate"
	"github.com/rjeczalik/netxtest"
)

func myLimitedListener(l net.Listener, limitGlobal, limitPerConn int) net.Listener {
	limited := tcprate.NewListener(l)
	limited.SetLimits(limitGlobal, limitPerConn)
	return limited
}

func main() {
	var test netxtest.LimitListenerTest

	test.RegisterFlags(flag.CommandLine)
	flag.Parse()

	if err := test.Run(myLimitedListener); err != nil {
		log.Fatal(err)
	}
}
