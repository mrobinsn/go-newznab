package main

var (
	apiBaseURL = ""
)

// Client is a type for interacting with usenet-crawler
type Client struct {
	apikey string
}

// New returns a new instance of Client
func New(apikey string) Client {
	return Client{
		apikey: apikey,
	}
}
