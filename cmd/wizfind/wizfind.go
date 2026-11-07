package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
)

// main runs wizfind
func main() {
	// Command line
	find := flag.Bool("f", false, "Find wiz lights")
	broadcastAddr := flag.String("b", "", "Broadcast address for wiz light discovery")
	set := flag.String("s", "", "Set wiz light with IP address specified in this option")
	color := flag.String("c", "#ff0000", "Color to set the wiz light to")
	flag.Parse()

	// If finding a light, do that
	if *find {
		os.Exit(findWizLights(*broadcastAddr))
	}

	// Or if setting a light, do that
	if *set != "" {
		os.Exit(setWizLight(*set, *color))
	}

	// If neither finding nor setting a light, show usage and exit with error
	flag.Usage()
	os.Exit(1)

}

// findWizLights discovers wiz lights on the network using the specified broadcast address
func findWizLights(broadcastAddr string) int {

	disc := wiz.Discover().
		WithCallback(discoveryCallback).
		EndAfter(5 * time.Second).
		BroadcastInterval(1 * time.Second)

	if broadcastAddr != "" {
		isValid := net.ParseIP(broadcastAddr) != nil
		if !isValid {
			fmt.Fprintf(os.Stderr, "Invalid broadcast address: %s\n", broadcastAddr)
			return 1
		}
		disc = disc.BroadcastAddress(broadcastAddr)

	}

	devs, err := disc.GetDiscoveredDevices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering devices: %v\n", err)
		return 1
	}
	if len(devs) == 0 {
		fmt.Printf("No devices found\n")
	} else {
		fmt.Printf("Found %d devices:\n", len(devs))
	}
	return 0
}

// discoveryCallback is called for each discovered wiz device
func discoveryCallback(dev *wiz.Device) {
	fmt.Printf("Discovered device, MAC: %s, IP: %s\n", dev.MAC(), dev.IP())
}

// setWizLight sets the color of a wiz light at the specified IP address
func setWizLight(ip string, color string) int {

	if ip == "" {
		fmt.Fprintf(os.Stderr, "Must provide ip address with -ip option\n")
		return 1
	}

	isValid := net.ParseIP(ip) != nil
	if !isValid {
		fmt.Fprintf(os.Stderr, "IP address invalid\n")
		return 1
	}

	settings, err := wiz.SettingsFromString(color)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Color %q is not valid\n", color)
		return 1
	}

	fmt.Printf("Setting light at %q to %q\n", ip, color)

	_, err = wiz.MustDevice(ip).SetPilot(settings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting light: %v\n", err)
		return 1
	}
	fmt.Printf("Done.\n")

	return 0
}
