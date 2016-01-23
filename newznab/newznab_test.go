package newznab

import (
	"crypto/md5"
	"encoding/base64"
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUsenetCrawlerClient(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	Convey("I have setup a torznab client", t, func() {
		apiKey := "YOUR_API_KEY_HERE"
		So(apiKey, ShouldNotEqual, "YOUR_API_KEY_HERE")
		client := New("path to your jackett api", apiKey, true)

		Convey("I can search using simple query", func() {
			results, err := client.SearchWithQuery(CategoryTVHD, "Supernatural S11E01")
			for _, result := range results {
				log.Info(result.JSONString())
			}

			So(err, ShouldBeNil)
			So(len(results), ShouldBeGreaterThan, 0)
		})
	})

	Convey("I have setup a nzb client", t, func() {
		apiKey := "YOUR_API_KEY_HERE"
		So(apiKey, ShouldNotEqual, "YOUR_API_KEY_HERE")
		client := New("https://usenet-crawler.com/api", apiKey, true)

		Convey("I can search using tvrage id", func() {
			results, err := client.SearchWithTVRage(CategoryTVSD, 2870, 10, 1)

			for _, result := range results {
				log.Info(result.JSONString())
			}

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
}
