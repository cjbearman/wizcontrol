package rpc

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/calendar"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

type Holiday struct {
	Name         string `json:"name"`
	DateThisYear string `json:"date_this_year"`
	daynum       int
}

// HolidaysResponse represents a mapping of holiday names to their respective dates this year.
type HolidaysResponse []Holiday

// HandleHolidaysRequest handles a request for holiday dates, returning a JSON-encoded mapping of holiday names to their respective dates this year.
func HandleHolidaysRequest(cnf *config.Config) []byte {
	logger := common.Logger.GetLogger("holiday-request")
	resp := HolidaysResponse{}

	for name, dayFunc := range calendar.HolidayNames {
		dayThisYear, err := calendar.DayThisYearFromDayNumber(dayFunc())
		if err != nil {
			// unlikely
			logger.Warnf("Failed to calculate holiday date this year for %s: %v", name, err)
			continue
		}

		resp = append(resp, Holiday{
			Name:         name,
			DateThisYear: strings.ToLower(dayThisYear),
			daynum:       dayFunc(),
		})
	}

	// Sort resp by daynum
	slices.SortFunc(resp, func(i, j Holiday) int {
		return i.daynum - j.daynum
	})

	jResp, _ := json.Marshal(resp)
	return jResp
}
