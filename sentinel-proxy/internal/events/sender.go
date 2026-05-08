package events

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func SendEvent(event SecurityEvent) {

	eventMap := map[string]interface{}{
		"event_type":  event.EventType,
		"request_id":  event.RequestID,
		"user_id":     event.User,
		"ip":          event.IP,
		"path":        event.Path,
		"method":      event.Method,
		"query":       event.Query,
		"attack_type": event.AttackType,
		"action":      event.Action,
		"timestamp":   event.Timestamp,
	}

	jsonData, err := json.Marshal(eventMap)
	if err != nil {
		log.Printf("JSON marshal failed: %v", err)
		return
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Post(
		"http://host.docker.internal:7777/event",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		log.Printf("Rust agent unreachable: %v", err)
		return
	}

	defer resp.Body.Close()

	log.Printf("Event shipped to Rust agent for user: %s", event.User)
}
