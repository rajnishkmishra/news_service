package news_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"news_service/models/vm"
	"news_service/services/llm_service"
	"news_service/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type NewsService struct {
	db         *gorm.DB
	elastic    *utils.Elastic
	llmService *llm_service.LlmService
}

func NewNewsService(db *gorm.DB, elastic *utils.Elastic, llmService *llm_service.LlmService) *NewsService {
	return &NewsService{
		db:         db,
		elastic:    elastic,
		llmService: llmService,
	}
}

func (n *NewsService) GetNewsByCategory(ctx *utils.Context, request vm.FetchNewsRequest) (response vm.NewsResponse, werr utils.WrapperError) {
	if strings.Trim(request.Category, " ") == "" {
		logrus.WithContext(ctx.Ctx).Error("invalid category")
		werr = utils.NewWrapperError(http.StatusBadRequest, errors.New("invalid category"))
		return
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"category": request.Category,
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"publication_date": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}

	elasticResponse, err := n.elastic.FetchFromElastic(ctx, query, utils.NEWS_INDEX, request.PaginationRequest)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, errors.New("something went wrong"))
		return
	}

	err = n.mapResponse(ctx, elasticResponse, request, &response)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, err)
		return
	}

	return
}

func (n *NewsService) GetNewsByScore(ctx *utils.Context, request vm.FetchNewsRequest) (response vm.NewsResponse, werr utils.WrapperError) {
	if request.Score <= 0 {
		logrus.WithContext(ctx.Ctx).Error("invalid score")
		werr = utils.NewWrapperError(http.StatusBadRequest, errors.New("invalid score"))
		return
	}
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"relevance_score": map[string]interface{}{
					"gte": request.Score,
				},
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"relevance_score": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}

	elasticResponse, err := n.elastic.FetchFromElastic(ctx, query, utils.NEWS_INDEX, request.PaginationRequest)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, errors.New("something went wrong"))
		return
	}

	err = n.mapResponse(ctx, elasticResponse, request, &response)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, err)
		return
	}
	return
}

func (n *NewsService) GetNewsBySource(ctx *utils.Context, request vm.FetchNewsRequest) (response vm.NewsResponse, werr utils.WrapperError) {
	if strings.Trim(request.Source, " ") == "" {
		logrus.WithContext(ctx.Ctx).Error("invalid source")
		werr = utils.NewWrapperError(http.StatusBadRequest, errors.New("invalid source"))
		return
	}
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_phrase": map[string]interface{}{
				"source_name": request.Source,
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"publication_date": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}

	elasticResponse, err := n.elastic.FetchFromElastic(ctx, query, utils.NEWS_INDEX, request.PaginationRequest)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, errors.New("something went wrong"))
		return
	}

	err = n.mapResponse(ctx, elasticResponse, request, &response)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, err)
		return
	}
	return
}

func (n *NewsService) GetNewsByLocation(ctx *utils.Context, request vm.FetchNewsRequest) (response vm.NewsResponse, werr utils.WrapperError) {
	if request.Lat <= 0 {
		logrus.WithContext(ctx.Ctx).Error("invalid latitute")
		werr = utils.NewWrapperError(http.StatusBadRequest, errors.New("invalid latitute"))
		return
	}
	if request.Long <= 0 {
		logrus.WithContext(ctx.Ctx).Error("invalid longitude")
		werr = utils.NewWrapperError(http.StatusBadRequest, errors.New("invalid longitude"))
		return
	}
	if request.Radius <= 0 {
		request.Radius = 10
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"geo_distance": map[string]interface{}{
				"distance": fmt.Sprintf("%vkm", request.Radius),
				"location": map[string]float64{
					"lat": request.Lat,
					"lon": request.Long,
				},
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"_geo_distance": map[string]interface{}{
					"location": map[string]float64{
						"lat": request.Lat,
						"lon": request.Long,
					},
					"order":           "asc",
					"unit":            "km",
					"distance_type":   "arc",
					"ignore_unmapped": true,
				},
			},
		},
	}

	elasticResponse, err := n.elastic.FetchFromElastic(ctx, query, utils.NEWS_INDEX, request.PaginationRequest)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, errors.New("something went wrong"))
		return
	}

	err = n.mapResponse(ctx, elasticResponse, request, &response)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, err)
		return
	}

	return
}

func (n *NewsService) GetNewsBySearch(ctx *utils.Context, request vm.FetchNewsRequest) (response vm.NewsResponse, werr utils.WrapperError) {
	if strings.Trim(request.Query, " ") == "" {
		logrus.WithContext(ctx.Ctx).Error("invalid query")
		werr = utils.NewWrapperError(http.StatusBadRequest, errors.New("invalid query"))
		return
	}

	llmOutput, err := n.llmService.AnalyzeQuery(ctx, request.Query)
	if err != nil {
		return
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"function_score": map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"should": append(
							[]interface{}{
								map[string]interface{}{
									"multi_match": map[string]interface{}{
										"query": request.Query,
										"fields": []string{
											"title^3",
											"description",
										},
										"type":      "best_fields",
										"fuzziness": "AUTO",
									},
								},
							},
							subQueriesBasedOnIntent(llmOutput)...,
						),
						"minimum_should_match": 1,
					},
				},

				"boost_mode": "sum",
				"score_mode": "sum",

				"functions": []interface{}{
					map[string]interface{}{
						"field_value_factor": map[string]interface{}{
							"field":    "relevance_score",
							"factor":   5,
							"modifier": "sqrt",
							"missing":  0,
						},
					},
				},
			},
		},
	}

	elasticResponse, err := n.elastic.FetchFromElastic(ctx, query, utils.NEWS_INDEX, request.PaginationRequest)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, err)
		return
	}

	err = n.mapResponse(ctx, elasticResponse, request, &response)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, err)
		return
	}

	return
}

func (n *NewsService) GetTrendingNewsByLocation(ctx *utils.Context, request vm.FetchNewsRequest) (response vm.NewsResponse, werr utils.WrapperError) {
	if request.Lat <= 0 {
		logrus.WithContext(ctx.Ctx).Error("invalid latitute")
		werr = utils.NewWrapperError(http.StatusBadRequest, errors.New("invalid latitute"))
		return
	}
	if request.Long <= 0 {
		logrus.WithContext(ctx.Ctx).Error("invalid longitude")
		werr = utils.NewWrapperError(http.StatusBadRequest, errors.New("invalid longitude"))
		return
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"function_score": map[string]interface{}{
				"boost_mode": "sum",
				"score_mode": "sum",
				"functions": []interface{}{
					map[string]interface{}{
						"field_value_factor": map[string]interface{}{
							"field":   "recent_activity_score",
							"factor":  1.0,
							"missing": 0,
						},
					},
					map[string]interface{}{
						"gauss": map[string]interface{}{
							"location": map[string]interface{}{
								"origin": map[string]float64{
									"lat": request.Lat,
									"lon": request.Long,
								},
								"scale":  "50km",
								"offset": "0km",
								"decay":  0.5,
							},
						},
					},
					map[string]interface{}{
						"gauss": map[string]interface{}{
							"last_event_time": map[string]interface{}{
								"origin": time.Now().Format(time.RFC3339),
								"scale":  "24h",
								"decay":  0.5,
							},
						},
					},
				},
			},
		},
	}

	elasticResponse, err := n.elastic.FetchFromElastic(ctx, query, utils.NEWS_INDEX, request.PaginationRequest)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, err)
		return
	}

	err = n.mapResponse(ctx, elasticResponse, request, &response)
	if err != nil {
		werr = utils.NewWrapperError(http.StatusInternalServerError, err)
		return
	}

	return
}

func (n *NewsService) mapResponse(ctx *utils.Context, elasticResponse map[string]interface{}, request vm.FetchNewsRequest, response *vm.NewsResponse) (err error) {
	hits, ok := elasticResponse["hits"].(map[string]interface{})
	if !ok {
		logrus.WithContext(ctx.Ctx).Errorf("error from elastic: %v", elasticResponse)
		err = errors.New("something went wrong")
		return
	}
	totalHits := int64(hits["total"].(map[string]interface{})["value"].(float64))
	data := make([]vm.News, 0)
	descriptions := make([]string, 0)
	for _, hit := range elasticResponse["hits"].(map[string]interface{})["hits"].([]interface{}) {
		source := hit.(map[string]interface{})["_source"]
		bytes, _ := json.Marshal(source)
		newsElastic := vm.NewsElastic{}
		json.Unmarshal(bytes, &newsElastic)
		data = append(data, vm.News{
			Title:           newsElastic.Title,
			Description:     newsElastic.Description,
			Url:             newsElastic.Url,
			PublicationDate: newsElastic.PublicationDate,
			SourceName:      newsElastic.SourceName,
			Category:        newsElastic.Category,
			RelevanceScore:  newsElastic.RelevanceScore,
			Latitude:        newsElastic.Location.Lat,
			Longitude:       newsElastic.Location.Lon,
		})
		descriptions = append(descriptions, newsElastic.Description)
	}

	llmSummary, err := n.llmService.GenerateSummary(ctx, descriptions)
	if err != nil {
		logrus.WithContext(ctx.Ctx).Error(err)
		err = nil
	}

	if len(data) == len(llmSummary) {
		for i := 0; i < len(data); i++ {
			data[i].LLMSummary = llmSummary[i]
		}
	}

	totalPage := totalHits / request.GetLimit()
	if totalHits%request.GetLimit() > 0 {
		totalPage++
	}
	response.Articles = data
	response.MetaResponse.PageNumber = request.GetPageNumber()
	response.MetaResponse.TotalPages = totalPage
	response.MetaResponse.TotalRecord = totalHits
	requestInBytes, _ := json.Marshal(request)
	var searchQuery map[string]interface{}
	json.Unmarshal(requestInBytes, &searchQuery)
	response.MetaResponse.SearchQuery = searchQuery
	return
}

func subQueriesBasedOnIntent(llmOutput *llm_service.LlmOutput) []interface{} {
	subQuery := make([]interface{}, 0)
	for _, i := range llmOutput.Intent {
		switch i {
		case "category":
			subQuery = append(subQuery, map[string]interface{}{
				"terms": map[string]interface{}{
					"category": llmOutput.Entities,
				},
			})
		case "source":
			subQuery = append(subQuery, map[string]interface{}{
				"terms": map[string]interface{}{
					"source_name.keyword": llmOutput.Entities,
				},
			})
		}
	}
	return subQuery
}
