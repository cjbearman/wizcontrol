package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
)

func main() {
	find := flag.Bool("f", false, "Find wiz lights")
	broadcastAddr := flag.String("b", "", "Broadcast address for wiz light discovery")

	set := flag.String("s", "", "Set wiz light with IP address specified in this option")
	color := flag.String("c", "#ff0000", "Color to set the wiz light to")

	flag.Parse()

	if *find {
		os.Exit(findWizLights(*broadcastAddr))
	}
	if *set != "" {
		os.Exit(setWizLight(*set, *color))
	}

	flag.Usage()
	os.Exit(1)

}

func findWizLights(broadcastAddr string) int {

	disc := wiz.Discover(false).
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

func discoveryCallback(dev *wiz.Device) {
	fmt.Printf("Discovered device, MAC: %s, IP: %s\n", macReformat(dev.GetMac()), dev.IP())
}

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

func macReformat(mac string) string {
	if len(mac) != 12 {
		return fmt.Sprintf("%s (invalid)", mac)
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		mac[0:2], mac[2:4], mac[4:6], mac[6:8], mac[8:10], mac[10:12])
}
