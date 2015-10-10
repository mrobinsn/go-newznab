package usenetcrawler

import (
	"encoding/json"
	"time"
)

// NZB represents an NZB found on the index
type NZB struct {
	ID          string
	Title       string
	Description string
	Size        int64
	AirDate     time.Time
	PubDate     time.Time
	NumGrabs    int
	NumComments int
	Comments    []Comment
}

// Comment represents a user comment left on an NZB record
type Comment struct {
	Title   string
	Content string
	PubDate time.Time
}

// JSONString returns a JSON string representation of this NZB
func (n NZB) JSONString() string {
	jsonString, _ := json.MarshalIndent(n, "", "  ")
	return string(jsonString)
}

// JSONString returns a JSON string representation of this Comment
func (c Comment) JSONString() string {
	jsonString, _ := json.MarshalIndent(c, "", "  ")
	return string(jsonString)
}
