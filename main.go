package main

import (
	"net"
	"os"

	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/support"
	"github.com/ninjasphere/sphere-go-led-controller/remote"
)

var log = logger.GetLogger("weather-pane")

func main() {
	pane := main.NewWeatherPane()

	tcpAddr, err := net.ResolveTCPAddr("tcp", config.String("localhost", "app-weather-pane.host")+":3115")
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	matrix := remote.NewMatrix(pane, conn)

	support.WaitUntilSignal()
}
