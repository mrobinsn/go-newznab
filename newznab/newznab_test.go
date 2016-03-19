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
		client := New(ts.URL+"/api", apiKey, true)

		Convey("I can search using simple query", func() {
			results, err := client.SearchWithQuery(CategoryTVHD, "Supernatural S11E01", "tvshows")
			//for _, result := range results {
			//	log.Info(result.JSONString())
			//}

			So(err, ShouldBeNil)
			So(len(results), ShouldBeGreaterThan, 0)
		})
	})

	Convey("I have setup a nzb client", t, func() {
		client := New(ts.URL+"/api", apiKey, false)

		Convey("Handle errors", func() {
			Convey("Return an error for an invalid search", func() {
				_, err := client.SearchWithTVDB(CategoryTVSD, 1234, 9, 2)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("I can get TV show information", func() {

			Convey("I can search using TheTVDB id", func() {
				results, err := client.SearchWithTVDB(CategoryTVSD, 75682, 10, 1)

				So(err, ShouldBeNil)
				So(len(results), ShouldBeGreaterThan, 0)
			})

			Convey("I can search using tvrage id", func() {
				results, err := client.SearchWithTVRage(CategoryTVSD, 2870, 10, 1)

				//for _, result := range results {
				//	log.Info(result.JSONString())
				//}

				So(err, ShouldBeNil)
				So(len(results), ShouldBeGreaterThan, 0)

				Convey("I can populate the comments for an NZB", func() {
					nzb := results[1]
					So(len(nzb.Comments), ShouldEqual, 0)
					So(nzb.NumComments, ShouldBeGreaterThan, 0)
					err := client.PopulateComments(&nzb)
					So(err, ShouldBeNil)

					for _, comment := range nzb.Comments {
						log.Info(comment.JSONString())
					}

					So(len(nzb.Comments), ShouldBeGreaterThan, 0)

					Convey("I can get the download url", func() {
						url := client.NZBDownloadURL(results[0])
						So(len(url), ShouldBeGreaterThan, 0)
						log.Infof("URL: %s", url)

						Convey("I can download the NZB", func() {
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
			})
		})
		Convey("I can get movie information", func() {

			Convey("I can search using an IMDB id", func() {
				results, err := client.SearchWithIMDB(CategoryMovieHD, "0364569")

				So(err, ShouldBeNil)
				So(len(results), ShouldBeGreaterThan, 0)
				Convey("Get movie specific fields from results", func() {
					Convey("IMDB ID", func() {
						imdbAttr := results[0].IMDBID

						So(imdbAttr, ShouldEqual, "0364569")
					})
					Convey("IMDB Title", func() {
						imdbAttr := results[0].IMDBTitle

						So(imdbAttr, ShouldEqual, "Oldboy")
					})
					Convey("IMDB Year", func() {
						imdbAttr := results[0].IMDBYear

						So(imdbAttr, ShouldEqual, 2003)
					})
					Convey("IMDB Score", func() {
						imdbAttr := results[0].IMDBScore

						So(imdbAttr, ShouldEqual, 8.4)
					})
					Convey("IMDB Cover URL", func() {
						imdbAttr := results[0].CoverURL

						So(imdbAttr, ShouldEqual, "https://dognzb.cr/content/covers/movies/thumbs/364569.jpg")
					})
				})
			})
		})
	})
}
