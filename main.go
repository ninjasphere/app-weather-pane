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

	// Create our pane. Must implement (github.com/ninjasphere/go-ninja/remote).pane
	pane := main.NewWeatherPane()

	// Connect to the led controller remote pane interface (port 3115)
	tcpAddr, err := net.ResolveTCPAddr("tcp", config.String("localhost", "led.host")+":3115")
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

	// Export our pane over this interface
	matrix := remote.NewMatrix(pane, conn)

	support.WaitUntilSignal()
}
