package backends

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"news_service/migration"
	"news_service/models"
	"news_service/models/vm"
	"news_service/services/apis"
	"news_service/services/llm_service"
	"news_service/services/news_service"
	"news_service/utils"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var r *gin.Engine
var buff chan models.UserActivity
var syncMap sync.Map

func PathHandler(backends utils.Backends) {
	migrationService := migration.NewMigrationService(backends.MySQLConn, backends.EsClient)
	go migrationService.InitMigration()
	r = backends.GinEngine

	elastic := utils.NewElastic(backends.EsClient)
	llmService := llm_service.NewLlmService()
	newsService := news_service.NewNewsService(backends.MySQLConn.DB, elastic, llmService)
	apis.NewNewsController(r, newsService)

	// readDataFromFile(backends)
	// createNewUsers(backends)
	// go updateElasticIndex(backends)
	// go updateElasticIndex(backends)
	// go updateElasticIndex(backends)
	// generateUserActivityEvents(backends)
}

func readDataFromFile(backends utils.Backends) {
	file, err := os.Open("news_data.json")
	if err != nil {
		logrus.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	newsRawArray := []vm.NewsRaw{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&newsRawArray); err != nil {
		log.Fatalf("Failed to decode JSON: %s", err)
	}

	dbNewsArray := []models.News{}
	for _, newsRaw := range newsRawArray {
		t, _ := time.Parse("2006-01-02T15:04:05", newsRaw.PublicationDate)
		dbNewsArray = append(dbNewsArray, models.News{
			Title:           newsRaw.Title,
			Description:     newsRaw.Title,
			Url:             newsRaw.Url,
			PublicationDate: t,
			SourceName:      newsRaw.SourceName,
			Category:        strings.Join(newsRaw.Category, ","),
			RelevanceScore:  newsRaw.RelevanceScore,
			Latitude:        newsRaw.Latitude,
			Longitude:       newsRaw.Longitude,
		})
	}

	err = backends.MySQLConn.DB.CreateInBatches(&dbNewsArray, 100).Error
	if err != nil {
		logrus.Fatal(err)
		return
	}

	for _, dbNews := range dbNewsArray {
		newsElastic := vm.NewsElastic{
			ID:              dbNews.ID,
			Title:           dbNews.Title,
			Description:     dbNews.Description,
			Url:             dbNews.Url,
			PublicationDate: dbNews.PublicationDate,
			SourceName:      dbNews.SourceName,
			Category:        strings.Split(dbNews.Category, ","),
			RelevanceScore:  dbNews.RelevanceScore,
			Location: vm.LatLon{
				Lat: dbNews.Latitude,
				Lon: dbNews.Longitude,
			},
		}

		data, _ := json.Marshal(newsElastic)
		res, err := backends.EsClient.Index(
			"news",
			bytes.NewReader(data),
			backends.EsClient.Index.WithContext(context.Background()),
		)
		if err != nil {
			logrus.Fatalf("Error indexing document: %s", err)
		}
		defer res.Body.Close()
	}
}

func createNewUsers(backends utils.Backends) {
	users := make([]models.User, 0)
	for i := 1; i <= 10; i++ {
		users = append(users, models.User{
			Name:        fmt.Sprintf("Rajnish %v", i),
			PhoneNumber: "1234567890",
			EmailID:     "raj@gmail.com",
		})
	}

	backends.MySQLConn.DB.Save(&users)
}

func generateUserActivityEvents(backends utils.Backends) {
	buff = make(chan models.UserActivity)
	go generateUserActivity(backends)
	go generateUserActivity(backends)
	go generateUserActivity(backends)
}

func generateUserActivity(backends utils.Backends) {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		activity := models.UserActivity{
			UserID:       uint64(rand.Int63n(10) + 1),
			NewsID:       uint64(rand.Int63n(2000) + 1),
			ActivityType: randomEventType(),
			ActivityTime: time.Now(),
		}
		saveActivityToDB(backends, activity)
	}
}

func updateElasticIndex(backends utils.Backends) {
	batchSize := 1
	for activity := range buff {
		val, _ := syncMap.LoadOrStore(activity.NewsID, []models.UserActivity{})
		userActivity := val.([]models.UserActivity)
		userActivity = append(userActivity, activity)
		syncMap.Store(activity.NewsID, userActivity)
		if len(userActivity) >= batchSize {
			go processAndFlush(backends, activity.NewsID)
			syncMap.Store(activity.NewsID, []models.UserActivity{})
		}
	}
}
func processAndFlush(backends utils.Backends, newsID uint64) {
	type Activity struct {
		Clicks        int
		Views         int
		Shares        int
		Likes         int
		LastEventTime time.Time
	}

	var result Activity

	err := backends.MySQLConn.DB.Raw(`
        SELECT 
            SUM(CASE WHEN activity_type = 'click' THEN 1 ELSE 0 END) AS clicks,
            SUM(CASE WHEN activity_type = 'view' THEN 1 ELSE 0 END) AS views,
            SUM(CASE WHEN activity_type = 'share' THEN 1 ELSE 0 END) AS shares,
			SUM(CASE WHEN activity_type = 'like' THEN 1 ELSE 0 END) AS likes,
            MAX(activity_time) AS last_event_time
        FROM user_activities
        WHERE news_id = ?
    `, newsID).Scan(&result).Error
	if err != nil {
		log.Fatal(err)
		return
	}

	hoursSinceEvent := time.Since(result.LastEventTime).Hours()
	lambda := 0.03
	recencyFactor := math.Exp(-lambda * hoursSinceEvent)
	trendingScore := float64(result.Likes*4 + result.Shares*3 + result.Views*2 + result.Clicks*1)
	trendingScore = trendingScore * recencyFactor

	type UpdateScriptParams struct {
		Script struct {
			Source string                 `json:"source"`
			Params map[string]interface{} `json:"params"`
		} `json:"script"`
		Query map[string]interface{} `json:"query"`
	}

	params := UpdateScriptParams{
		Query: map[string]interface{}{
			"term": map[string]interface{}{
				"id": map[string]interface{}{
					"value": newsID,
				},
			},
		},
	}
	params.Script.Source = "ctx._source.recent_activity_score = params.recent_activity_score; ctx._source.last_event_time = params.last_event_time;"
	params.Script.Params = map[string]interface{}{
		"recent_activity_score": trendingScore,
		"last_event_time":       result.LastEventTime.Format(time.RFC3339),
	}

	body, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}

	res, err := backends.EsClient.UpdateByQuery(
		[]string{"news"},
		backends.EsClient.UpdateByQuery.WithBody(bytes.NewReader(body)),
		backends.EsClient.UpdateByQuery.WithContext(context.Background()),
	)

	if err != nil {
		log.Fatal(err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatal("something went wrong")
		return
	}
}

func randomEventType() string {
	choices := []string{"like", "share", "click", "view"}
	return choices[rand.Intn(len(choices))]
}

func saveActivityToDB(backends utils.Backends, activity models.UserActivity) {
	err := backends.MySQLConn.DB.Save(&activity).Error
	if err != nil {
		log.Fatal(err)
		return
	}
	buff <- activity
}
