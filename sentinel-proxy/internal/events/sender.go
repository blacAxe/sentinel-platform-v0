package events

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func SendEvent(event SecurityEvent) {

    // Create a map to add extra context for LumenLog
    eventMap := make(map[string]interface{})
    
    eventMap["request_id"] = event.RequestID
    eventMap["ip"] = event.IP
    eventMap["path"] = event.Path
    eventMap["action"] = event.Action
    eventMap["attack_type"] = event.AttackType
    
    // Add placeholder for user 
    eventMap["user_id"] = "anonymous" 

    jsonData, _ := json.Marshal(eventMap)

	resp, err := http.Post(
		"http://localhost:9001/events",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return
	}
	defer resp.Body.Close()
}
