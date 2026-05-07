package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/omar/sentinel-proxy/internal/storage"
	"github.com/omar/sentinel-proxy/internal/events"
)

var LogChan chan string

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	IP        string `json:"ip"`
	Path      string `json:"path"`
	Query     string `json:"query"`
	Action    string `json:"action"`
	Reason    string `json:"reason"`
}

var logFile *os.File

func Init() {
	f, err := os.OpenFile("data/logs.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	logFile = f
}

func Log(entry LogEntry) {
	entry.Timestamp = time.Now().Format(time.RFC3339)

	data, _ := json.Marshal(entry)
	logFile.Write(append(data, '\n'))

    msg := fmt.Sprintf(
        `{"ip":"%s","path":"%s","query":"%s","action":"%s","reason":"%s"}`,
        entry.IP,
        entry.Path,
        entry.Query,
        entry.Action,
        entry.Reason,
    )

    _, err := db.DB.Exec(`
        INSERT INTO logs (timestamp, ip, path, query, action, reason)
        VALUES (?, ?, ?, ?, ?, ?)`,
        entry.Timestamp,
        entry.IP,
        entry.Path,
        entry.Query,
        entry.Action,
        entry.Reason,
    )

    if err != nil {
        log.Println("DB insert error:", err)
    }

    // send to dashboard
    select {
    case LogChan <- msg:
    default:
    }
    
}

func LogEvent(event events.SecurityEvent) {
    // write to file
    data, _ := json.Marshal(event)
    logFile.Write(append(data, '\n'))

    // send JSON to dashboard
    msgBytes, _ := json.Marshal(event)
    msg := string(msgBytes)

    select {
    case LogChan <- msg:
    default:
    }
}