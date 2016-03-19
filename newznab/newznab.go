package newznab

import (
	"crypto/tls"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	// CategoryTVHD is the category for high-definition TV shows
	CategoryTVHD = 5040
	// CategoryTVSD is the category for standard-definition TV shows
	CategoryTVSD = 5030
	// CategoryMovieHD is the category for high-definition TV shows
	CategoryMovieHD = 2040
	// CategoryMovieSD is the category for standard-definition TV shows
	CategoryMovieSD = 2030
)

// Client is a type for interacting with a newznab or torznab api
type Client struct {
	apikey     string
	apiBaseURL string
	client     *http.Client
}

// New returns a new instance of Client
func New(baseURL string, apikey string, insecure bool) Client {
	ret := Client{
		apikey:     apikey,
		apiBaseURL: baseURL,
	}
	if insecure {
		ret.client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}}
	} else {
		ret.client = &http.Client{}
	}
	return ret
}

// SearchWithTVRage returns NZBs for the given parameters
func (c Client) SearchWithTVRage(category int, tvRageID int, season int, episode int) ([]NZB, error) {
	return c.search(url.Values{
		"rid":     []string{strconv.Itoa(tvRageID)},
		"cat":     []string{strconv.Itoa(category)},
		"season":  []string{strconv.Itoa(season)},
		"episode": []string{strconv.Itoa(episode)},
		"t":       []string{"tvsearch"},
	})
}

// SearchWithIMDB returns NZBs for the given parameters
func (c Client) SearchWithIMDB(category int, imdbID string) ([]NZB, error) {
	return c.search(url.Values{
		"imdbid": []string{imdbID},
		"cat":    []string{strconv.Itoa(category)},
		"t":      []string{"movie"},
	})
}

// SearchWithQuery returns NZBs for the given parameters
func (c Client) SearchWithQuery(category int, query string, searchType string) ([]NZB, error) {
	return c.search(url.Values{
		"q":   []string{query},
		"cat": []string{strconv.Itoa(category)},
		"t":   []string{searchType},
	})
}

func (c Client) search(vals url.Values) ([]NZB, error) {
	vals.Set("apikey", c.apikey)
	//vals.Set("t", "tvsearch")
	var nzbs []NZB
	log.Debug("newznab:Client:Search: searching")
	resp, err := c.getURL(c.buildURL(vals))
	if err != nil {
		return nzbs, err
	}
	var feed SearchResponse
	err = xml.Unmarshal(resp, &feed)
	if err != nil {
		return nil, err
	}
	log.WithField("num", len(feed.Channel.NZBs)).Info("newznab:Client:Search: found NZBs")
	for _, gotNZB := range feed.Channel.NZBs {
		nzb := NZB{
			Title:          gotNZB.Title,
			Description:    gotNZB.Description,
			PubDate:        gotNZB.Date.Add(0),
			DownloadURL:    gotNZB.Enclosure.URL,
			SourceEndpoint: c.apiBaseURL,
			SourceAPIKey:   c.apikey,
		}

		for _, attr := range gotNZB.Attributes {
			switch attr.Name {
			case "tvairdate":
				if parsedAirDate, err := time.Parse(time.RFC1123Z, attr.Value); err != nil {
					log.Errorf("newznab:Client:Search: failed to parse date: %v: %v", attr.Value, err)
				} else {
					nzb.AirDate = parsedAirDate
				}
			case "guid":
				nzb.ID = attr.Value
			case "size":
				parsedInt, _ := strconv.ParseInt(attr.Value, 0, 64)
				nzb.Size = parsedInt
			case "grabs":
				parsedInt, _ := strconv.ParseInt(attr.Value, 0, 32)
				nzb.NumGrabs = int(parsedInt)
			case "comments":
				parsedInt, _ := strconv.ParseInt(attr.Value, 0, 32)
				nzb.NumComments = int(parsedInt)
			case "seeders":
				parsedInt, _ := strconv.ParseInt(attr.Value, 0, 32)
				nzb.Seeders = int(parsedInt)
				nzb.IsTorrent = true
			case "peers":
				parsedInt, _ := strconv.ParseInt(attr.Value, 0, 32)
				nzb.Peers = int(parsedInt)
				nzb.IsTorrent = true
			case "infohash":
				nzb.InfoHash = attr.Value
				nzb.IsTorrent = true
			default:
				log.WithFields(log.Fields{
					"name":  attr.Name,
					"value": attr.Value,
				}).Debug("encounted unknown attribute")
			}
		}
		if nzb.Size == 0 {
			nzb.Size = gotNZB.Size
		}
		nzbs = append(nzbs, nzb)
	}
	return nzbs, nil
}

// PopulateComments fills in the Comments for the given NZB
func (c Client) PopulateComments(nzb *NZB) error {
	log.Debug("newznab:Client:PopulateComments: getting comments")
	data, err := c.getURL(c.buildURL(url.Values{
		"t":      []string{"comments"},
		"id":     []string{nzb.ID},
		"apikey": []string{c.apikey},
	}))
	if err != nil {
		return err
	}
	var resp commentResponse
	err = xml.Unmarshal(data, &resp)
	if err != nil {
		return err
	}

	for _, rawComment := range resp.Channel.Comments {
		comment := Comment{
			Title:   rawComment.Title,
			Content: rawComment.Description,
		}
		if parsedPubDate, err := time.Parse(time.RFC1123Z, rawComment.PubDate); err != nil {
			log.WithFields(log.Fields{
				"pub_date": rawComment.PubDate,
				"err":      err,
			}).Error("newznab:Client:PopulateComments: failed to parse date")
		} else {
			comment.PubDate = parsedPubDate
		}
		nzb.Comments = append(nzb.Comments, comment)
	}
	return nil
}

// NZBDownloadURL returns a URL to download the NZB from
func (c Client) NZBDownloadURL(nzb NZB) string {
	return c.buildURL(url.Values{
		"t":      []string{"get"},
		"id":     []string{nzb.ID},
		"apikey": []string{c.apikey},
	})
}

// DownloadNZB returns the bytes of the actual NZB file for the given NZB
func (c Client) DownloadNZB(nzb NZB) ([]byte, error) {
	return c.getURL(c.NZBDownloadURL(nzb))
}

func (c Client) getURL(url string) ([]byte, error) {
	log.WithField("url", url).Debug("newznab:Client:getURL: getting url")
	res, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}

	var data []byte
	data, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	log.WithField("num_bytes", len(data)).Debug("newznab:Client:getURL: retrieved")

	return data, nil
}

func (c Client) buildURL(vals url.Values) string {
	parsedURL, err := url.Parse(c.apiBaseURL)
	if err != nil {
		log.WithError(err).Error("failed to parse base API url")
		return ""
	}

	parsedURL.RawQuery = vals.Encode()
	return parsedURL.String()
}

type commentResponse struct {
	Channel struct {
		Comments []rssComment `xml:"item"`
	} `xml:"channel"`
}

type rssComment struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}
