package validator

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/cjbearman/wiz/pkg/wiz"
)

var componentNameRegexp = regexp.MustCompile(`^[a-z0-9_\-]+$`)
var macReg1 = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
var macReg2 = regexp.MustCompile(`^[a-zA-Z0-9]{12}$`)
var macReg3 = regexp.MustCompile(`^([0-9A-Fa-f]{4}\.){2}([0-9A-Fa-f]{4})$`)

func ComponentName(name string) (string, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "", errors.New("missing")
	}
	ok := componentNameRegexp.MatchString(name)
	if !ok {
		return "", errors.New("must contain a-z, 0-9, _, or - only")
	}

	return name, nil
}

func RequiredComponentName(name string) (string, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "", errors.New("missing")
	}
	return ComponentName(name)
}

func RequiredColorString(color string) (string, error) {
	color = strings.TrimSpace(strings.ToLower(color))
	if color == "" {
		return "", errors.New("missing")
	}

	colorstr, err := wiz.RationalizeColor(color)
	if err != nil {
		return "", fmt.Errorf("%q is not a valid color string", color)
	}

	return colorstr, nil

}

func IP(ip string, required bool) (string, error) {
	ip = strings.TrimSpace(ip)
	if !required && ip == "" {
		return "", nil
	}
	if net.ParseIP(ip) == nil {
		return "", errors.New("invalid IP address")
	}
	return ip, nil
}

func MAC(mac string, required bool) (string, error) {
	mac = strings.TrimSpace(strings.ToLower(mac))
	if !required && mac == "" {
		return "", nil
	}

	if macReg1.MatchString(mac) {
		return mac, nil
	}
	if macReg2.MatchString(mac) {
		return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
			mac[0:2], mac[2:4], mac[4:6], mac[6:8], mac[8:10], mac[10:12]), nil
	}
	if macReg3.MatchString(mac) {
		return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
			mac[0:2], mac[2:4], mac[5:7], mac[7:9], mac[10:12], mac[12:14]), nil
	}

	return "", errors.New("XX:XX:XX:XX:XX:XX or XXXX.XXXX.XXXX or XXXXXXXXXXXX")
}
