package usenetcrawler

import (
	"crypto/md5"
	"encoding/base64"
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

var apiKey = "YOUR_API_KEY_HERE"

func TestUsenetCrawlerClient(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	Convey("I have setup a client", t, func() {
		So(apiKey, ShouldNotEqual, "YOUR_API_KEY_HERE")
		client := New(apiKey)

		Convey("I can search", func() {
			results, err := client.Search(CategoryTVSD, 2870, 10, 1)

			for _, result := range results {
				log.Infoln(result.Pretty())
			}

			So(err, ShouldBeNil)
			So(len(results), ShouldBeGreaterThan, 0)

			Convey("I can populate the comments for an NZB", func() {
				nzb := results[0]
				So(len(nzb.Comments), ShouldEqual, 0)
				So(nzb.NumComments, ShouldBeGreaterThan, 0)
				err := client.PopulateComments(&nzb)
				So(err, ShouldBeNil)

				for _, comment := range nzb.Comments {
					log.Infoln(comment.Pretty())
				}

				So(len(nzb.Comments), ShouldBeGreaterThan, 0)
			})

			Convey("I can get the download url", func() {
				url := client.DownloadURL(results[0])
				So(len(url), ShouldBeGreaterThan, 0)
				log.Infof("URL: %s", url)
			})

			Convey("I can download the NZB", func() {
				bytes, err := client.Download(results[0])
				So(err, ShouldBeNil)

				md5Sum := md5.Sum(bytes)
				log.Infof("Downloaded %d bytes, md5: %s", len(bytes), base64.StdEncoding.EncodeToString(md5Sum[:]))

				So(len(bytes), ShouldBeGreaterThan, 0)
			})
		})
	})
}
