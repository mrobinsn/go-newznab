# go-newznab

> newznab XML API client for Go (golang)

## Documentation
https://godoc.org/github.com/tehjojo/go-newznab/newznab

## Features
- Search for episode with TVRage ID
- Search for files with category and query
- Get comments for a NZB
- Get NZB download URL
- Download NZB

## Installation
To install the package run `go get github.com/tehjojo/go-newznab`
To use it in your application, import `github.com/tehjojo/go-newznab/newznab`

## Library Usage
```
client := newznab.New("http://my-usenet-indexer/api", "my-api-key", false)
results, _ := client.SearchWithTVRage(newznab.CategoryTVHD, 35048, 3, 1)
```

## Contributing
Pull requests welcome.
