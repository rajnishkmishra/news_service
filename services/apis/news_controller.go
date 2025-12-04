package apis

import (
	"news_service/services/news_service"
	"news_service/utils"

	"github.com/gin-gonic/gin"
)

const (
	GetNewsByCategory         = "/news/catorgory/:category"
	GetNewsByScore            = "/news/score/:score"
	GetNewsBySearch           = "/news/search"
	GetNewsBySource           = "/news/source/:source"
	GetNewsByNearBy           = "/news/nearby"
	GetTrendingNewsByLocation = "/trending"
)

func NewNewsController(engine *gin.Engine, newsService *news_service.NewsService) {
	router := engine.Group("/api/v1")
	router.GET(GetNewsByCategory, utils.Controller(utils.NewOptions(newsService.GetNewsByCategory)))
	router.GET(GetNewsByScore, utils.Controller(utils.NewOptions(newsService.GetNewsByScore)))
	router.GET(GetNewsBySearch, utils.Controller(utils.NewOptions(newsService.GetNewsBySearch)))
	router.GET(GetNewsBySource, utils.Controller(utils.NewOptions(newsService.GetNewsBySource)))
	router.GET(GetNewsByNearBy, utils.Controller(utils.NewOptions(newsService.GetNewsByLocation)))
	router.GET(GetTrendingNewsByLocation, utils.Controller(utils.NewOptions(newsService.GetTrendingNewsByLocation)))
}
