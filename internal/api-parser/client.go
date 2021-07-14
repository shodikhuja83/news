package apiParser

import "net/http"

type ParserClient struct {
	client *http.Client
}

func CreateParserClient(httpClient *http.Client) *ParserClient {
	return &ParserClient{client: httpClient}
}
