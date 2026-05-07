package rules

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Rule struct {
	Name     string `json:"name"`
	Pattern  string `json:"pattern"`
	Field    string `json:"field"`
	Priority int    `json:"priority"`
}

var Rules []Rule

func SetRules(r []Rule) {
	Rules = r
}

func LoadRules() {
	filePaths := []string{
		"internal/rules/rules.json",
		"../rules/rules.json",
		"./rules.json",
	}

	var data []byte
	var err error

	for _, path := range filePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		panic("could not load rules.json")
	}

	json.Unmarshal(data, &Rules)
}

func EvaluateRequest(r *http.Request, query string) (bool, string) {
	query = strings.ToLower(query)
	path := strings.ToLower(r.URL.Path)

	for _, rule := range Rules {
		var target string

		switch rule.Field {
		case "query":
			target = query
		case "path":
			target = path
		default:
			target = query // fallback
		}

		matched, err := regexp.MatchString(rule.Pattern, target)
		if err != nil {
			continue // skip bad regex instead of crashing
		}

		if matched {
			return true, rule.Name
		}
	}

	return false, ""
}
