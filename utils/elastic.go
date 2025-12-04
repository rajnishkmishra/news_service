package utils

import (
	"bytes"
	"encoding/json"
	"news_service/models/vm"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

const (
	NEWS_INDEX = "news"
)

type Elastic struct {
	esClient *elasticsearch.Client
}

func NewElastic(esClient *elasticsearch.Client) *Elastic {
	return &Elastic{
		esClient: esClient,
	}
}

func (e *Elastic) FetchFromElastic(ctx *Context, query map[string]interface{},
	indexName string, paginationRequest vm.PaginationRequest) (response map[string]interface{}, err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		logrus.Fatalf("Error encoding query: %s", err)
		return
	}

	searchRes, err := e.esClient.Search(
		e.esClient.Search.WithContext(ctx.Ctx),
		e.esClient.Search.WithIndex(indexName),
		e.esClient.Search.WithBody(&buf),
		e.esClient.Search.WithTrackTotalHits(true),
		e.esClient.Search.WithFrom(int(paginationRequest.GetLimit()*(paginationRequest.GetPageNumber()-1))),
		e.esClient.Search.WithSize(int(paginationRequest.GetLimit())),
	)
	if err != nil {
		logrus.WithContext(ctx.Ctx).Error(err)
		return
	}
	defer searchRes.Body.Close()

	response = make(map[string]interface{})
	if err = json.NewDecoder(searchRes.Body).Decode(&response); err != nil {
		logrus.WithContext(ctx.Ctx).Error(err)
		return
	}
	return
}
