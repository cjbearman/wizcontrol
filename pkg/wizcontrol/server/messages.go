package server

type DayRequest struct {
	Scenes []string `json:"scenes"`
}
type LightSetRequest struct {
	Color   string `json:"color"`
	LockFor int    `json:"for"`
	Force   bool   `json:"force"`
}

type LightRespItem struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
}
