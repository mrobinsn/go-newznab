package newznab

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUsenetCrawlerClient(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	apiKey := "gibberish"

	// Set up our mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var f []byte
		var err error

		reg := regexp.MustCompile(`\W`)
		fixedPath := reg.ReplaceAllString(r.URL.RawQuery, "_")

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
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("File not found"))
		} else {
			w.Write(f)
		}
	}))

	defer ts.Close()

	t.Run("torznab client", func(t *testing.T) {
		client := New(ts.URL, apiKey, 1234, true)

		t.Run("Simple query search", func(t *testing.T) {
			categories := []int{CategoryTVHD}
			results, err := client.SearchWithQuery(categories, "Supernatural S11E01", "tvshows")
			require.NoError(t, err)
			require.NotEmpty(t, results, "expected results")
		})
	})

	t.Run("nzb client", func(t *testing.T) {
		client := New(ts.URL, apiKey, 1234, false)
		categories := []int{CategoryTVSD}

		t.Run("invalid search", func(t *testing.T) {
			_, err := client.SearchWithTVDB(categories, 1234, 9, 2)
			require.Error(t, err, "expected an error")
		})

		t.Run("invalid api usage", func(t *testing.T) {
			_, err := client.SearchWithTVDB(categories, 5678, 9, 2)
			require.Error(t, err, "expected an error")
			require.EqualError(t, err, "newznab api error 100: Invalid API Key")
		})

		t.Run("valid category and TheTVDB id", func(t *testing.T) {
			results, err := client.SearchWithTVDB(categories, 75682, 10, 1)
			require.NoError(t, err)
			require.NotEmpty(t, results, "expected results")
		})

		t.Run("valid category and TVMaze id", func(t *testing.T) {
			results, err := client.SearchWithTVMaze(categories, 65, 10, 1)
			require.NoError(t, err)
			require.NotEmpty(t, results, "expected results")
		})

		t.Run("valid category and tvrage id", func(t *testing.T) {
			results, err := client.SearchWithTVRage(categories, 2870, 10, 1)
			require.NoError(t, err)
			require.NotEmpty(t, results, "expected results")

			t.Run("populate comments", func(t *testing.T) {
				nzb := results[1]
				require.Empty(t, nzb.Comments)
				require.NotZero(t, nzb.NumComments)
				err := client.PopulateComments(&nzb)
				require.NoError(t, err)
				require.NotEmpty(t, nzb.Comments, "expected at least one comment")
				for _, comment := range nzb.Comments {
					require.NotEmpty(t, comment, "comment should not be empty")
				}
			})

			t.Run("download url", func(t *testing.T) {
				url, err := client.NZBDownloadURL(results[0])
				require.NoError(t, err)
				require.NotEmpty(t, url, "expected a url")
			})

			t.Run("download nzb", func(t *testing.T) {
				bytes, err := client.DownloadNZB(results[0])
				require.NoError(t, err)
				require.NotEmpty(t, bytes, "expected to download something")
			})
		})

		t.Run("multiple categories and IMDB id", func(t *testing.T) {
			cats := []int{
				CategoryMovieHD,
				CategoryMovieBluRay,
			}
			results, err := client.SearchWithIMDB(cats, "0371746")
			require.NoError(t, err)
			require.NotEmpty(t, results, "expected results")

			require.Equal(t, "2040", results[0].Category[1])
			require.Equal(t, "2050", results[22].Category[1])
		})

		t.Run("single category and IMDB id", func(t *testing.T) {
			cats := []int{CategoryMovieHD}
			results, err := client.SearchWithIMDB(cats, "0364569")
			require.NoError(t, err)
			require.NotEmpty(t, results, "expected results")

			t.Run("movie specific fields", func(t *testing.T) {
				require.Equal(t, "0364569", results[0].IMDBID)
				require.Equal(t, "Oldboy", results[0].IMDBTitle)
				require.Equal(t, 2003, results[0].IMDBYear)
				require.Equal(t, float32(8.4), results[0].IMDBScore)
				require.Equal(t, "https://dognzb.cr/content/covers/movies/thumbs/364569.jpg", results[0].CoverURL)
			})
		})

		t.Run("recent items via RSS", func(t *testing.T) {
			num := 50
			categories := []int{CategoryMovieAll, CategoryTVAll}

			t.Run("recent items", func(t *testing.T) {
				results, err := client.LoadRSSFeed(categories, num)
				require.NoError(t, err)
				require.Len(t, results, num)
				require.Equal(t, "bcdbf3f1e7a1ef964527f1d40d5ec639", results[0].ID)
				require.Equal(t, "030517-VSHS0101720WDA20H264V", results[6].Title)

				t.Run("airdate with RFC1123Z format", func(t *testing.T) {
					require.Equal(t, 2017, results[7].AirDate.Year())
				})

				t.Run("usenetdate with RFC3339 format", func(t *testing.T) {
					require.Equal(t, 2017, results[7].UsenetDate.Year())
				})
			})

			t.Run("up until", func(t *testing.T) {
				results, err := client.LoadRSSFeedUntilNZBID(categories, num, "29527a54ac54bb7533abacd7dad66a6a", 0)
				require.NoError(t, err)
				require.Len(t, results, 101)

				t.Run("boundary results", func(t *testing.T) {
					require.Equal(t, "8841b21c4d2fb96f0d47ca24cae9a5b7", results[0].ID)
					require.Equal(t, "2c6c0e2ac562db69d8b3646deaf2d0cd", results[len(results)-1].ID)
				})

				t.Run("RSS up until with failures/retries", func(t *testing.T) {
					results, err := client.LoadRSSFeedUntilNZBID(categories, num, "does-not-exist", 2)
					require.NoError(t, err)
					require.Len(t, results, 100)
				})
			})
		})
	})
}
