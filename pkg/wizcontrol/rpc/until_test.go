package rpc

import (
	"testing"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

func TestUntilParseBasic(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"2024-06-15T14:00:00Z", true},
		{"Mon, 15 Jun 2024 14:00:00 UTC", true},
		{"15 Jun 24 14:00 UTC", true},
		{"invalid", false},
	}

	for _, test := range tests {
		_, err := untilParse(test.input)
		if (err == nil) != test.expected {
			t.Errorf("untilParse(%q) = %v, expected success: %v", test.input, err, test.expected)
		}
	}
}

func TestUntilParseTimeOnly(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"14:00:00", true},
		{"02:00 PM", true},
		{"1:03", true},
		{"0:00", true},
		{"invalid", false},
	}

	for _, test := range tests {
		ts, err := untilParse(test.input)
		if (err == nil) != test.expected {
			t.Errorf("untilParse(%q) = %v, expected success: %v", test.input, err, test.expected)
			continue
		}
		if err == nil {
			// Time must be in the future
			if ts.Before(time.Now().In(common.Location)) {
				t.Errorf("untilParse(%q) returned past time: %v", test.input, ts)
			}

		}

	}
}
