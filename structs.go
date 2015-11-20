package usenetcrawler

import (
	"encoding/json"
	"time"
)

// NZB represents an NZB found on the index
type NZB struct {
	ID          string    `json:"id,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Size        int64     `json:"size,omitempty"`
	AirDate     time.Time `json:"air_date,omitempty"`
	PubDate     time.Time `json:"pub_date,omitempty"`
	NumGrabs    int       `json:"num_grabs,omitempty"`
	NumComments int       `json:"num_comments,omitempty"`
	Comments    []Comment `json:"comments,omitempty"`
}

// Comment represents a user comment left on an NZB record
type Comment struct {
	Title   string    `json:"title,omitempty"`
	Content string    `json:"content,omitempty"`
	PubDate time.Time `json:"pub_date,omitempty"`
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
