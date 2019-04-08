# go-newznab
[![GoDoc](https://godoc.org/github.com/mrobinsn/go-newznab/newznab?status.svg)](https://godoc.org/github.com/mrobinsn/go-newznab/newznab)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrobinsn/go-newznab)](https://goreportcard.com/report/github.com/mrobinsn/go-newznab)
[![Build Status](https://travis-ci.org/mrobinsn/go-newznab.svg?branch=master)](https://travis-ci.org/mrobinsn/go-newznab)
[![Coverage Status](https://coveralls.io/repos/github/mrobinsn/go-newznab/badge.svg?branch=master)](https://coveralls.io/github/mrobinsn/go-newznab?branch=master)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)


> newznab/torznab XML API client for Go (golang)

## Documentation
[GoDoc](https://godoc.org/github.com/mrobinsn/go-newznab/newznab)

## Features
- TV and Movie search
- Search for files with category(s) and query
- Get comments for a NZB
- Get NZB download URL
- Download NZB
- Get latest releases via RSS

## Installation
To install the package run `go get github.com/mrobinsn/go-newznab`
To use it in your application, import `github.com/mrobinsn/go-newznab/newznab`

## Library Usage

### Initialize a client:
```
client := newznab.New("http://my-usenet-indexer", "my-api-key", 1234, false)

```
Note the missing `/api` part of the URL. Depending on the called method either `/api` or `/rss` will be appended to the given base URL. A valid user ID is only required for RSS methods.

### Get the capabilities of your tracker
```
caps, _ := client.Capabilities()
```
You will want to check the result of this to determine if your tracker supports searching by tvrage, imdb, tvmaze, etc.

### Search using a tvrage id:
```
categories := []int{
    newznab.CategoryTVHD,
    newznab.CategoryTVSD,
}
results, _ := client.SearchWithTVRage(categories, 35048, 3, 1)
```

### Search using an imdb id:
```
categories := []int{
    newznab.CategoryMovieHD,
    newznab.CategoryMovieBluRay,
}
results, _ := client.SearchWithIMDB(categories, "0364569")
```

### Search using a tvmaze id:
```
categories := []int{
    newznab.CategoryTVHD,
    newznab.CategoryTVSD,
}
results, _ := client.SearchWithTVMaze(categories, 80, 3, 1)
```

### Search using a name and set of categories:
```
results, _ := client.SearchWithQueries(categories, "Oldboy", "movie")
```

### Get latest releases for set of categories:
```
results, _ := client.SearchWithQuery(categories, "", "movie")
```

### Load latest releases via RSS:
```
results, _ := client.LoadRSSFeed(categories, 50)
```

### Load latest releases via RSS up to a given NZB id:
```
results, _ := client.LoadRSSFeedUntilNZBID(categories, 50, "nzb-guid", 15)
```

## Contributing
Pull requests welcome.
