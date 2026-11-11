package validator

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/cjbearman/wiz/pkg/wiz"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

var componentNameRegexp = regexp.MustCompile(`^[a-z0-9_\-]+$`)
var macReg1 = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
var macReg2 = regexp.MustCompile(`^[a-zA-Z0-9]{12}$`)
var macReg3 = regexp.MustCompile(`^([0-9A-Fa-f]{4}\.){2}([0-9A-Fa-f]{4})$`)

// ComponentName validates all component names as lower case strings containing alphanumerics, - and _ only.
// empty component names are allowed
func ComponentName(name string) (string, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return name, nil
	}
	ok := componentNameRegexp.MatchString(name)
	if !ok {
		return "", errors.New("must contain a-z, 0-9, _, or - only")
	}

	return name, nil
}

// RequiredComponentName validates that a component name is provided and follows the same rules as ComponentName.
func RequiredComponentName(name string) (string, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "", errors.New("missing")
	}
	return ComponentName(name)
}

// RequiredColorString validates that a color string is provided and is a valid color.
func RequiredColorString(color string) (string, error) {
	color = strings.TrimSpace(strings.ToLower(color))
	if color == "" {
		return "", errors.New("missing")
	}

	colorstr, err := wiz.RationalizeColor(color)
	if err != nil {
		return "", fmt.Errorf("%q is not a valid color string", color)
	}

	return strings.ToLower(colorstr), nil

}

// IP validates that the provided string is a valid IP address. If required is true, an empty string is considered invalid.
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

// MAC validates that the provided string is a valid MAC address. If required is true, an empty string is considered invalid.
// Returns a normalized MAC address in xx:xx:xx:xx:xx:xx format or error.
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

type StringValidator func(string) (string, error)

func ValidateAnyString(s string) (string, error) {
	return s, nil
}

func ValidateIP(ip string) (string, error) {
	if net.ParseIP(ip) == nil {
		return "", errors.New("invalid IP address")
	}
	return ip, nil
}

func ValidateMAC(mac string) (string, error) {
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

var IDReg = regexp.MustCompile(`^[a-z0-9\-_]+$`)

func ValidateID(id string) (string, error) {
	if !IDReg.MatchString(id) {
		return "", errors.New("must contain only lowercase letters, numbers, - and _")
	}
	return id, nil
}

func ValidateFor(forStr string) (string, error) {
	_, err := common.ForParser(forStr)
	if err != nil {
		return "", err
	}
	return forStr, nil
}

func ValidateDuration(durationStr string) (string, error) {
	parsed, err := time.ParseDuration(durationStr)
	if err != nil {
		return "", errors.New("invalid duration")
	}
	if parsed <= 0 {
		return "", errors.New("duration must be positive")
	}
	return durationStr, nil
}

func ValidateStr(inbound string, outbound *string, validator StringValidator, required bool) error {
	inbound = strings.TrimSpace(inbound)
	if inbound == "" {
		if required {
			return errors.New("required field missing")
		}
		*outbound = ""
		return nil
	}

	validated, err := validator(inbound)
	if err != nil {
		return err
	}
	*outbound = validated
	return nil
}
