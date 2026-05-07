package events

type SecurityEvent struct {
	EventType      string `json:"event_type"`
	RequestID      string `json:"request_id"`
	User 		   string `json:"user"`
	IP             string `json:"ip"`
	Path           string `json:"path"`
	Method         string `json:"method"`
	Query          string `json:"query"`
	AttackDetected bool   `json:"attack_detected"`
	AttackType     string `json:"attack_type"`
	Action         string `json:"action"`
	Timestamp      int64  `json:"timestamp"`
}
