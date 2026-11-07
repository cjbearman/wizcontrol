package rpc

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

// GeneralResponse represents a standard response for RPC requests, indicating success or failure and an optional error message.
type GeneralResponse struct {
	Success bool    `json:"success"`
	Error   *string `json:"error,omitempty"`
}

// NewGeneralResponse creates a new GeneralResponse with the given success status and error message.
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

// RpcError represents an error that occurred during an RPC request, including an error message and an HTTP status code.
type RpcError struct {
	error  string
	status int
}

// Error returns the error message associated with the RpcError.
func (e *RpcError) Error() string {
	return e.error
}

// Status returns the HTTP status code associated with the RpcError.
func (e *RpcError) Status() int {
	return e.status
}

// NewRpcError creates a new RpcError with the given HTTP status code and formatted error message.
// 'status' is the HTTP status code, 'format' is a format string, and 'args' are the arguments for the format string.
func NewRpcError(status int, format string, args ...interface{}) *RpcError {
	return &RpcError{
		error:  fmt.Sprintf(format, args...),
		status: status,
	}
}

// validateForUntil validates the 'for' and 'until' parameters, ensuring that exactly one is provided and returns the corresponding time.
// 'forspec' is the 'for' duration string, 'until' is the 'until' time string.
// Returns the calculated time and an optional RpcError if the input is invalid.
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

// forParser parses the 'for' duration string and returns the corresponding time.
// 'forSpec' is the 'for' duration string.
// Returns the calculated time and an optional RpcError if the input is invalid.
// Accepts any duration or "day", "week", "month"
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

// parseableTimeDateFormats lists the time and date formats that can be parsed by untilParse.
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

// untilParse parses the 'until' time string and returns the corresponding time.
// 'until' is the 'until' time string.
// Returns the calculated time and an optional RpcError if the input is invalid.
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
