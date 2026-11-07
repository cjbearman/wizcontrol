package rpc

import (
	"encoding/json"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

// Time represents a specific time of day with hour and minute components.
type Time struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// Calendar represents a calendar with its associated metadata and selection criteria.
type Calendar struct {
	Name      string `json:"name"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	Before    *Time  `json:"before,omitempty"`
	After     *Time  `json:"after,omitempty"`
	Mode      string `json:"mode"`
	Frequency string `json:"frequency,omitempty"`
}

// CalendarsResponse represents the response structure for a calendars request, containing a list of calendars.
type CalendarsResponse struct {
	Calendars []Calendar `json:"calendars"`
}

// HandleCalendarsRequest handles a request for calendar data, returning a JSON-encoded response containing the list of calendars.
func HandleCalendarsRequest(cnf *config.Config) []byte {

	resp := &CalendarsResponse{}
	for _, calendar := range cnf.Calendars {

		var before, after *Time
		if calendar.Before() != nil {
			before = &Time{
				Hour:   calendar.Before().Hour,
				Minute: calendar.Before().Minute,
			}
		}
		if calendar.After() != nil {
			after = &Time{
				Hour:   calendar.After().Hour,
				Minute: calendar.After().Minute,
			}
		}

		resp.Calendars = append(resp.Calendars,
			Calendar{
				Name:      calendar.Name(),
				From:      calendar.OriginalFrom(),
				To:        calendar.OriginalTo(),
				Before:    before,
				After:     after,
				Mode:      calendar.Selector().ModeStr(),
				Frequency: calendar.Selector().FrequencyStr(),
			})
	}
	jResp, _ := json.Marshal(resp)
	return jResp
}
