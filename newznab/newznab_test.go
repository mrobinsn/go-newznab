package newznab

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUsenetCrawlerClient(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	apiKey := "gibberish"

	// Set up our mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var f []byte
		var err error

		reg := regexp.MustCompile(`\W`)
		fixedPath := reg.ReplaceAllString(r.URL.RawQuery, "_")

		log.Info("Local fixture path: tests/fixtures" + r.URL.Path + "/" + fixedPath)

		if r.URL.Query()["t"][0] == "get" {
			// Fetch nzb
			nzbID := r.URL.Query()["id"][0]
			filePath := fmt.Sprintf("../tests/fixtures/nzbs/%v.nzb", nzbID)
			f, err = ioutil.ReadFile(filePath)
		} else {
			// Get xml
			filePath := fmt.Sprintf("../tests/fixtures%v/%v.xml", r.URL.Path, fixedPath)
			f, err = ioutil.ReadFile(filePath)
		}

		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("File not found"))
		} else {
			w.Write(f)
		}
	}))

	defer ts.Close()

	Convey("I have setup a torznab client", t, func() {
		client := New(ts.URL, apiKey, 1234, true)

		Convey("I can search using simple query", func() {
			categories := []int{CategoryTVHD}
			results, err := client.SearchWithQuery(categories, "Supernatural S11E01", "tvshows")
			//for _, result := range results {
			//	log.Info(result.JSONString())
			//}

			So(err, ShouldBeNil)
			So(len(results), ShouldBeGreaterThan, 0)
		})
	})

	Convey("I have setup a nzb client", t, func() {
		client := New(ts.URL, apiKey, 1234, false)
		categories := []int{CategoryTVSD}

		Convey("Handle errors", func() {

			Convey("Return an error for an invalid search.", func() {
				_, err := client.SearchWithTVDB(categories, 1234, 9, 2)
				So(err, ShouldNotBeNil)
			})

			Convey("Return an error for invalid api usage.", func() {
				_, err := client.SearchWithTVDB(categories, 5678, 9, 2)
				So(err.Error(), ShouldEqual, "newznab api error 100: Invalid API Key")
			})
		})

		Convey("When getting TV show information", func() {

			Convey("Given a category and a TheTVDB id", func() {
				results, err := client.SearchWithTVDB(categories, 75682, 10, 1)

				Convey("A valid result is returned.", func() {
					So(err, ShouldBeNil)
					So(len(results), ShouldBeGreaterThan, 0)
				})
			})

			Convey("When given a category and a tvrage id", func() {
				results, err := client.SearchWithTVRage(categories, 2870, 10, 1)

				Convey("A valid result is returned.", func() {
					So(err, ShouldBeNil)
					So(len(results), ShouldBeGreaterThan, 0)
				})

				Convey("I can populate the comments for an NZB.", func() {
					nzb := results[1]
					So(len(nzb.Comments), ShouldEqual, 0)
					So(nzb.NumComments, ShouldBeGreaterThan, 0)
					err := client.PopulateComments(&nzb)
					So(err, ShouldBeNil)

					for _, comment := range nzb.Comments {
						log.Info(comment.JSONString())
					}

					So(len(nzb.Comments), ShouldBeGreaterThan, 0)
				})

				Convey("I can get the download url.", func() {
					url := client.NZBDownloadURL(results[0])
					So(len(url), ShouldBeGreaterThan, 0)
					log.Infof("URL: %s", url)
				})

				Convey("I can download the NZB.", func() {
					bytes, err := client.DownloadNZB(results[0])
					So(err, ShouldBeNil)

					md5Sum := md5.Sum(bytes)
					log.WithFields(log.Fields{
						"num_bytes": len(bytes),
						"md5":       base64.StdEncoding.EncodeToString(md5Sum[:]),
					}).Info("downloaded")

					So(len(bytes), ShouldBeGreaterThan, 0)
				})
			})
		})

		Convey("When getting movie information", func() {
			Convey("Given multiple categories and an IMDB id", func() {
				cats := []int{
					CategoryMovieHD,
					CategoryMovieBluRay,
				}
				results, err := client.SearchWithIMDB(cats, "0371746")

				So(err, ShouldBeNil)
				So(len(results), ShouldBeGreaterThan, 0)

				Convey("The results have different categories.", func() {
					So(results[0].Category[1], ShouldEqual, "2040")
					So(results[22].Category[1], ShouldEqual, "2050")
				})
			})

			Convey("Given a single category and an IMDB id", func() {
				cats := []int{CategoryMovieHD}
				results, err := client.SearchWithIMDB(cats, "0364569")

				So(err, ShouldBeNil)
				So(len(results), ShouldBeGreaterThan, 0)

				Convey("I can get movie specific fields", func() {

					Convey("An IMDB id.", func() {
						imdbAttr := results[0].IMDBID
						So(imdbAttr, ShouldEqual, "0364569")
					})

					Convey("An IMDB title.", func() {
						imdbAttr := results[0].IMDBTitle
						So(imdbAttr, ShouldEqual, "Oldboy")
					})

					Convey("An IMDB year.", func() {
						imdbAttr := results[0].IMDBYear
						So(imdbAttr, ShouldEqual, 2003)
					})

					Convey("An IMDB score.", func() {
						imdbAttr := results[0].IMDBScore
						So(imdbAttr, ShouldEqual, 8.4)
					})

					Convey("A cover URL.", func() {
						imdbAttr := results[0].CoverURL
						So(imdbAttr, ShouldEqual, "https://dognzb.cr/content/covers/movies/thumbs/364569.jpg")
					})
				})
			})
		})

		Convey("When getting recent items via RSS", func() {
			num := 50
			categories := []int{CategoryMovieAll, CategoryTVAll}

			Convey("I can load the current RSS feed.", func() {
				results, err := client.LoadRSSFeed(categories, num)

				Convey("A valid result is returned.", func() {
					So(err, ShouldBeNil)
					So(len(results), ShouldEqual, num)
				})

				Convey("A TV result is present.", func() {
					guid := results[0].ID
					So(guid, ShouldEqual, "bcdbf3f1e7a1ef964527f1d40d5ec639")
				})

				Convey("A Movie result is present.", func() {
					title := results[6].Title
					So(title, ShouldEqual, "030517-VSHS0101720WDA20H264V")
				})

				Convey("An airdate with RFC1123Z format is parsed.", func() {
					year := results[7].AirDate.Year()
					So(year, ShouldEqual, 2017)
				})

				Convey("An usenetdate with RFC3339 format is parsed.", func() {
					year := results[7].UsenetDate.Year()
					So(year, ShouldEqual, 2017)
				})

			})

			Convey("I can load the RSS feed up to a given NZB ID.", func() {
				results, err := client.LoadRSSFeedUntilNZBID(categories, num, "29527a54ac54bb7533abacd7dad66a6a", 0)

				Convey("A valid result is returned.", func() {
					So(err, ShouldBeNil)
					So(len(results), ShouldEqual, 101)
				})

				Convey("Everything up to the given ID is returned.", func() {
					firstID := results[0].ID
					So(firstID, ShouldEqual, "8841b21c4d2fb96f0d47ca24cae9a5b7")

					lastID := results[len(results)-1].ID
					So(lastID, ShouldEqual, "2c6c0e2ac562db69d8b3646deaf2d0cd")
				})
			})

			Convey("I can load the RSS feed up to a given NZB ID but will stop after N tries", func() {
				results, err := client.LoadRSSFeedUntilNZBID(categories, num, "does-not-exist", 2)

				Convey("100 results with 2 requests were fetched.", func() {
					So(err, ShouldBeNil)
					So(len(results), ShouldEqual, 100)
				})
			})
		})
	})
}
