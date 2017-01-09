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

## Installation
To install the package run `go get github.com/mrobinsn/go-newznab`
To use it in your application, import `github.com/mrobinsn/go-newznab/newznab`

## Library Usage

Initialize a client:
```
client := newznab.New("http://my-usenet-indexer/api", "my-api-key", false)

```

Search using a tvrage id:
```
categories := []int{
    newznab.CategoryTVHD,
    newznab.CategoryTVSD,
}
results, _ := client.SearchWithTVRage(categories, 35048, 3, 1)
```

Search using an imdb id:
```
categories := []int{
    newznab.CategoryMovieHD,
    newznab.CategoryMovieBluRay,
}
results, _ := client.SearchWithIMDB(categories, "0364569")
```

Search using a name and set of categories:
```
results, _ := client.SearchWithQueries(categories, "Oldboy", "movie")
```

Get latest releases for set of categories:
```
results, _ := client.SearchWithQuery(categories, "", "movie")
```

## Contributing
Pull requests welcome.
