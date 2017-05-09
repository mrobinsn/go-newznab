# go-newznab

> newznab XML API client for Go (golang)

## Documentation
https://godoc.org/github.com/mrobinsn/go-newznab/newznab

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
