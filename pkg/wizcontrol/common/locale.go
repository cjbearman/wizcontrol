package common

import "time"

// Location is set during config initialization to represent the local time zone of the system.
// It defaults to the local location of the system
var Location *time.Location = time.Now().Local().Location()
