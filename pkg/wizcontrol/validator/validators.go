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

var macReg1 = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
var macReg2 = regexp.MustCompile(`^[a-zA-Z0-9]{12}$`)
var macReg3 = regexp.MustCompile(`^([0-9A-Fa-f]{4}\.){2}([0-9A-Fa-f]{4})$`)

// StringVaildator is any function that takes a string, validates it and returns a possibly
// corrected version of that string or a validation error
type StringValidator func(string) (string, error)

// ValidateStr validates the inbound string using the provided StringValidator.
// If the string is required and missing, it returns an error.
// The validated string is assigned to the outbound pointer.
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

// Any is a placeholder validator that takes any string and returns it unmodified
func Any(s string) (string, error) {
	return s, nil
}

// IP validates that the input string is a valid IP address (IPv4 or IPv6)
func IP(ip string) (string, error) {
	if net.ParseIP(ip) == nil {
		return "", errors.New("invalid IP address")
	}
	return ip, nil
}

// MAC validates that the input string is a valid MAC address in one of the accepted formats
func MAC(mac string) (string, error) {
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

// ID validates that the input string contains only lowercase letters, numbers, - and _
func ID(id string) (string, error) {
	if !IDReg.MatchString(id) {
		return "", errors.New("must contain only lowercase letters, numbers, - and _")
	}
	return id, nil
}

// For validates that the input string is a valid "for" specification
func For(forStr string) (string, error) {
	_, err := common.ForParser(forStr)
	if err != nil {
		return "", err
	}
	return forStr, nil
}

// Duration validates that the input string is a valid duration and is positive
func Duration(durationStr string) (string, error) {
	parsed, err := time.ParseDuration(durationStr)
	if err != nil {
		return "", errors.New("invalid duration")
	}
	if parsed <= 0 {
		return "", errors.New("duration must be positive")
	}
	return durationStr, nil
}

// ColorValue validates that the input string is a valid color specification
func ColorValue(cval string) (string, error) {
	colorstr, err := wiz.RationalizeColor(cval)
	if err != nil {
		return "", fmt.Errorf("%q is not a valid color specification", cval)
	}

	return colorstr, nil
}

// BaseTopic validates that the input string is a valid base MQTT topic
func BaseTopic(topic string) (string, error) {
	if strings.ContainsAny(topic, "#+") {
		return "", errors.New("topic must not contain '#' or '+'")
	}
	if strings.HasPrefix(topic, "$") {
		return "", errors.New("topic must not start with '$'")
	}
	return topic, nil
}
