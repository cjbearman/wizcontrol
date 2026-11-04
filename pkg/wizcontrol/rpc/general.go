package rpc

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/lights"
)

// Will be set by MQTT (prevents import loop)
var LightStatusDispatcher func(*lights.Light) error

type GeneralResponse struct {
	Success bool    `json:"success"`
	Error   *string `json:"error,omitempty"`
}

func NewGeneralResponse(success bool, errmsg string) []byte {
	resp := GeneralResponse{
		Success: success,
	}
	if errmsg != "" {
		resp.Error = &errmsg
	}
	data, _ := json.Marshal(resp)
	return data
}

type RpcError struct {
	error  string
	status int
}

func (e *RpcError) Error() string {
	return e.error
}

func (e *RpcError) Status() int {
	return e.status
}

func NewRpcError(status int, format string, args ...interface{}) *RpcError {
	return &RpcError{
		error:  fmt.Sprintf(format, args...),
		status: status,
	}
}

func validateForUntil(forspec, until string) (time.Time, *RpcError) {
	if forspec == "" && until == "" {
		return time.Time{}, NewRpcError(400, "either 'for' or 'until' is required")
	}
	if forspec != "" && until != "" {
		return time.Time{}, NewRpcError(400, "'for' and 'until' are mutually exclusive")
	}
	if forspec != "" {
		return forParser(forspec)
	}
	if until != "" {
		return untilParse(until)
	}
	return time.Time{}, NewRpcError(400, "either 'for' or 'until' is required")
}

func forParser(forSpec string) (time.Time, *RpcError) {
	forSpec = strings.TrimSpace(strings.ToLower(forSpec))
	// Accept a standard Go duration string, e.g., "1h30m"
	duration, err := time.ParseDuration(forSpec)

	if err == nil {
		return time.Now().In(common.Location).Add(duration), nil
	}

	// Okay, so that wasn't it, we have a few other suported options
	if forSpec == "day" {
		now := time.Now().In(common.Location)
		return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, common.Location), nil
	}

	if forSpec == "week" {
		now := time.Now().In(common.Location)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		daysUntilEndOfWeek := 7 - weekday
		endOfWeek := now.AddDate(0, 0, daysUntilEndOfWeek)
		return time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 59, 59, 0, common.Location), nil
	}

	if forSpec == "month" {
		now := time.Now().In(common.Location)
		// Day zero of any next month rationally gives us the last day of the current month,
		// even though we might be setting a month of 13. Works.
		endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, common.Location)
		return endOfMonth, nil
	}

	return time.Time{}, NewRpcError(400, "invalid 'for' duration: %v", err)
}

var parseableTimeDateFormats = []string{
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.DateTime,
	time.DateOnly,
	time.TimeOnly,
	"3:04 PM",
	"15:04",
}

func untilParse(until string) (time.Time, *RpcError) {
	for _, fmt := range parseableTimeDateFormats {
		ts, err := time.ParseInLocation(fmt, until, common.Location)
		if err == nil {
			// If this was a time only, it might be in the past, so we need to adjust it to the future
			if ts.Before(time.Now().In(common.Location)) {
				ts = time.Date(
					time.Now().In(common.Location).Year(),
					time.Now().In(common.Location).Month(),
					time.Now().In(common.Location).Day(),
					ts.Hour(),
					ts.Minute(),
					ts.Second(),
					ts.Nanosecond(),
					common.Location,
				)

				// Now, that time could still be earlier in the current day, so we need to add 24 hours to move it to tomorrow
				// should that arise
				if ts.Before(time.Now().In(common.Location)) {
					ts = ts.Add(24 * time.Hour)
				}
			}

			return ts, nil
		}
	}
	return time.Time{}, NewRpcError(400, "invalid 'until' time: %v", until)
}
