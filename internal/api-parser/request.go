package apiParser

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

func (p *ParserClient) GetArticlesByTags(tags []string) (results map[string]Response) {
	results = make(map[string]Response)
	for _, tag := range tags {
		resp, err := p.client.Get("https://newsapi.org/v2/top-headlines?apiKey=536363846b4b48ea9d2ca6dcca90aa50&country=ru&category=" + tag)
		if err != nil {
			log.Errorf("error getting response for tag %s: %+v", tag, err)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("error reading response body for tag %s: %+v", tag, err)
			continue
		}

		var articlesResponse Response
		if err = json.Unmarshal(body, &articlesResponse); err != nil {
			log.Errorf("error marshalling response body for tag %s: %+v", tag, err)
			continue
		}
		results[tag] = articlesResponse
	}
	return
}
