package newznab

import (
	"crypto/tls"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fmt"

	log "github.com/Sirupsen/logrus"
)

// Various constants for categories
const (
	// TV Categories
	// CategoryTVAll is for all shows
	CategoryTVAll = 5000
	// CategoryTVForeign is for foreign shows
	CategoryTVForeign = 5020
	// CategoryTVSD is for standard-definition shows
	CategoryTVSD = 5030
	// CategoryTVHD is for high-definition shows
	CategoryTVHD = 5040
	// CategoryTVOther is for other shows
	CategoryTVOther = 5050
	// CategoryTVSport is for sports shows
	CategoryTVSport = 5060

	// Movie categories
	// CategoryMovieAll is for all movies
	CategoryMovieAll = 2000
	// CategoryMovieForeign is for foreign movies
	CategoryMovieForeign = 2010
	// CategoryMovieOther is for other movies
	CategoryMovieOther = 2020
	// CategoryMovieSD is for standard-definition movies
	CategoryMovieSD = 2030
	// CategoryMovieHD is for high-definition movies
	CategoryMovieHD = 2040
	// CategoryMovieBluRay is for blu-ray movies
	CategoryMovieBluRay = 2050
	// CategoryMovie3D is for 3-D movies
	CategoryMovie3D = 2060
)

// Client is a type for interacting with a newznab or torznab api
type Client struct {
	apikey     string
	apiBaseURL string
	apiUserID  int
	client     *http.Client
}

// New returns a new instance of Client
func New(baseURL string, apikey string, userID int, insecure bool) Client {
	ret := Client{
		apikey:     apikey,
		apiBaseURL: baseURL,
		apiUserID:  userID,
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
func (c Client) SearchWithTVRage(categories []int, tvRageID int, season int, episode int) ([]NZB, error) {
	return c.search(url.Values{
		"rid":     []string{strconv.Itoa(tvRageID)},
		"cat":     c.splitCats(categories),
		"season":  []string{strconv.Itoa(season)},
		"episode": []string{strconv.Itoa(episode)},
		"t":       []string{"tvsearch"},
	})
}

// SearchWithTVDB returns NZBs for the given parameters
func (c Client) SearchWithTVDB(categories []int, tvDBID int, season int, episode int) ([]NZB, error) {
	return c.search(url.Values{
		"tvdbid":  []string{strconv.Itoa(tvDBID)},
		"cat":     c.splitCats(categories),
		"season":  []string{strconv.Itoa(season)},
		"episode": []string{strconv.Itoa(episode)},
		"t":       []string{"tvsearch"},
	})
}

// SearchWithIMDB returns NZBs for the given parameters
func (c Client) SearchWithIMDB(categories []int, imdbID string) ([]NZB, error) {
	return c.search(url.Values{
		"imdbid": []string{imdbID},
		"cat":    c.splitCats(categories),
		"t":      []string{"movie"},
	})
}

// SearchWithQuery returns NZBs for the given parameters
func (c Client) SearchWithQuery(categories []int, query string, searchType string) ([]NZB, error) {
	return c.search(url.Values{
		"q":   []string{query},
		"cat": c.splitCats(categories),
		"t":   []string{searchType},
	})
}

// LoadRSSFeed returns up to <num> of the most recent NZBs of the given categories.
func (c Client) LoadRSSFeed(categories []int, num int) ([]NZB, error) {
	return c.rss(url.Values{
		"num": []string{strconv.Itoa(num)},
		"t":   c.splitCats(categories),
		"dl":  []string{"1"},
	})
}

// LoadRSSFeedUntilNZBID fetches NZBs until a given NZB id is reached.
func (c Client) LoadRSSFeedUntilNZBID(categories []int, num int, id string, maxRequests int) ([]NZB, error) {
	count := 0
	var nzbs []NZB
	for {
		partition, err := c.rss(url.Values{
			"num":    []string{strconv.Itoa(num)},
			"t":      c.splitCats(categories),
			"dl":     []string{"1"},
			"offset": []string{strconv.Itoa(num * count)},
		})
		count++
		if err != nil {
			return nil, err
		}
		for k, nzb := range partition {
			if nzb.ID == id {
				return append(nzbs, partition[:k]...), nil
			}
		}
		nzbs = append(nzbs, partition...)
		if maxRequests != 0 && count == maxRequests {
			break
		}
	}
	return nzbs, nil
}

func (c Client) splitCats(cats []int) []string {
	var categories, catsOut []string
	for _, v := range cats {
		categories = append(categories, strconv.Itoa(v))
	}
	catsOut = append(catsOut, strings.Join(categories, ","))
	return catsOut
}

func (c Client) rss(vals url.Values) ([]NZB, error) {
	vals.Set("r", c.apikey)
	vals.Set("i", strconv.Itoa(c.apiUserID))
	return c.process(vals, rssPath)
}

func (c Client) search(vals url.Values) ([]NZB, error) {
	vals.Set("apikey", c.apikey)
	return c.process(vals, apiPath)
}

func (c Client) process(vals url.Values, path string) ([]NZB, error) {
	var nzbs []NZB
	log.Debug("newznab:Client:Search: searching")
	resp, err := c.getURL(c.buildURL(vals, path))
	if err != nil {
		return nzbs, err
	}
	var feed SearchResponse
	err = xml.Unmarshal(resp, &feed)
	if err != nil {
		return nil, err
	}
	if feed.ErrorCode != 0 {
		return nil, fmt.Errorf("newznab api error %d: %s", feed.ErrorCode, feed.ErrorDesc)
	}
	log.WithField("num", len(feed.Channel.NZBs)).Debug("newznab:Client:Search: found NZBs")
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
				if parsedAirDate, err := parseDate(attr.Value); err != nil {
					log.Errorf("newznab:Client:Search: failed to parse tvairdate: %v", err)
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
			case "category":
				nzb.Category = append(nzb.Category, attr.Value)
			case "genre":
				nzb.Genre = attr.Value
			case "tvdbid":
				nzb.TVDBID = attr.Value
			case "rageid":
				nzb.TVRageID = attr.Value
			case "info":
				nzb.Info = attr.Value
			case "season":
				nzb.Season = attr.Value
			case "episode":
				nzb.Episode = attr.Value
			case "tvtitle":
				nzb.TVTitle = attr.Value
			case "rating":
				parsedInt, _ := strconv.ParseInt(attr.Value, 0, 32)
				nzb.Rating = int(parsedInt)
			case "imdb":
				nzb.IMDBID = attr.Value
			case "imdbtitle":
				nzb.IMDBTitle = attr.Value
			case "imdbyear":
				parsedInt, _ := strconv.ParseInt(attr.Value, 0, 32)
				nzb.IMDBYear = int(parsedInt)
			case "imdbscore":
				parsedFloat, _ := strconv.ParseFloat(attr.Value, 32)
				nzb.IMDBScore = float32(parsedFloat)
			case "coverurl":
				nzb.CoverURL = attr.Value
			case "usenetdate":
				if parsedUsetnetDate, err := parseDate(attr.Value); err != nil {
					log.Errorf("newznab:Client:Search: failed to parse usenetdate: %v", err)
				} else {
					nzb.UsenetDate = parsedUsetnetDate
				}
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
	}, apiPath))
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
	}, apiPath)
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

func (c Client) buildURL(vals url.Values, path string) string {
	parsedURL, err := url.Parse(c.apiBaseURL + path)
	if err != nil {
		log.WithError(err).Error("failed to parse base API url")
		return ""
	}

	parsedURL.RawQuery = vals.Encode()
	return parsedURL.String()
}

func parseDate(date string) (time.Time, error) {
	formats := []string{time.RFC3339, time.RFC1123Z}
	var parsedTime time.Time
	var err error
	for _, format := range formats {
		if parsedTime, err = time.Parse(format, date); err == nil {
			return parsedTime, nil
		}
	}
	return parsedTime, fmt.Errorf("failed to parse date %s as one of %s", date, strings.Join(formats, ", "))
}

const (
	apiPath = "/api"
	rssPath = "/rss"
)

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
