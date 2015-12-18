package main

import (
	"os"
	"testing"
)

var (
	outgoingPort = "10001"
	output       = make(chan string)
)

func TestMain(m *testing.M) {
	go outgoingServer(outgoingPort, output)
	ret := m.Run()
	os.Exit(ret)
}
