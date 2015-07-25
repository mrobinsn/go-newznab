package usenetcrawler

import (
	"fmt"
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

// Pretty returns a pretty-print string representation of this NZB
func (n NZB) Pretty() string {
	return fmt.Sprintf(`NZB(%s)
        Title: %s
        Description: %s
        Size: %d bytes
        AirDate: %s
        PubDate: %s
        NumGrabs: %d
        NumComments: %d
        `, n.ID, n.Title, n.Description, n.Size, n.AirDate,
		n.PubDate, n.NumGrabs, n.NumComments)
}

// Pretty returns a pretty-print string representation of this Comment
func (c Comment) Pretty() string {
	return fmt.Sprintf(`Comment
        Title: %s
        Content: %s
        PubDate: %s
        `, c.Title, c.Content, c.PubDate)
}
