package newznab

import (
	"encoding/json"
	"encoding/xml"
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

	// Torznab specific stuff
	Seeders   int
	Peers     int
	InfoHash  string
	IsTorrent bool
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

type SearchResponse struct {
	// RSS version of the response.
	Version string `xml:"version,attr"`
	Channel struct {
		Title string `xml:"title"`
		Link  struct {
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"http://www.w3.org/2005/Atom link"`
		Description string `xml:"description"`
		Language    string `xml:"language,omitempty"`
		Webmaster   string `xml:"webmaster,omitempty"`
		Category    string `xml:"category,omitempty"`
		Image       struct {
			URL         string `xml:"url"`
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description,omitempty"`
			Width       int    `xml:"width,omitempty"`
			Height      int    `xml:"height,omitempty"`
		} `xml:"image"`

		Response struct {
			Offset int `xml:"offset,attr"`
			Total  int `xml:"total,attr"`
		} `xml:"http://www.newznab.com/DTD/2010/feeds/attributes/ response"`

		// All NZBs that match the search query, up to the response limit.
		NZBs []RawNZB `xml:"item"`
	} `xml:"channel"`
}

// RawNZB represents a single NZB item in search results.
type RawNZB struct {
	Title    string `xml:"title,omitempty"`
	Link     string `xml:"link,omitempty"`
	Category struct {
		Domain string `xml:"domain,attr"`
		Value  string `xml:",chardata"`
	} `xml:"category,omitempty"`

	GUID struct {
		GUID        string `xml:",chardata"`
		IsPermaLink bool   `xml:"isPermaLink,attr"`
	} `xml:"guid,omitempty"`

	Comments    string `xml:"comments"`
	Description string `xml:"description"`
	Author      string `xml:"author,omitempty"`

	Source struct {
		URL   string `xml:"url,attr"`
		Value string `xml:"url,chardata"`
	} `xml:"source,omitempty"`

	Date Time `xml:"pubDate,omitempty"`

	Enclosure struct {
		URL    string `xml:"url,attr"`
		Length string `xml:"length,attr"`
		Type   string `xml:"type,attr"`
	} `xml:"enclosure,omitempty"`

	Attributes []struct {
		XMLName xml.Name
		Name    string `xml:"name,attr"`
		Value   string `xml:"value,attr"`
	} `xml:"attr"`
}

type Time struct {
	time.Time
}

func (t *Time) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeToken(start)
	e.EncodeToken(xml.CharData([]byte(t.UTC().Format(time.RFC822))))
	e.EncodeToken(xml.EndElement{start.Name})
	return nil
}

func (t *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw string

	err := d.DecodeElement(&raw, &start)
	if err != nil {
		return err
	}
	date, err := time.Parse(time.RFC1123Z, raw)

	if err != nil {
		return err
	}

	*t = Time{date}
	return nil

}
