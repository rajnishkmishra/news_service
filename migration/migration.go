package migration

import (
	"context"
	"fmt"
	"log"
	"news_service/models"
	"news_service/utils"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type Service struct {
	mySqlClient *utils.MySQLConn
	esClient    *elasticsearch.Client
}

func NewMigrationService(mySqlClient *utils.MySQLConn, esClient *elasticsearch.Client) Service {
	return Service{mySqlClient: mySqlClient, esClient: esClient}
}

func (m *Service) migrate(entity interface{}) (err error) {
	defer utils.Recovery()
	if db := m.mySqlClient.DB; db != nil {
		if err = db.AutoMigrate(entity); err != nil {
			log.Fatal("migration failed")
		}
	}
	return err
}

func (m *Service) InitMigration() {
	m.dbMigration()

	m.indexCreationAndMapping()
}

func (m *Service) dbMigration() {
	log.Printf("goroutine::DB table migration started...")
	dbTables := map[string]interface{}{
		"news":            &models.News{},
		"users":           &models.User{},
		"user_activities": &models.UserActivity{},
	}

	for tableName, table := range dbTables {
		err := m.migrate(table)
		if err != nil {
			log.Fatalf("migration failed for table: %v", tableName)
		}
	}
	log.Printf("goroutine::DB table migration concluded")
}

func (m *Service) indexCreationAndMapping() {
	createIndexRes, err := m.esClient.Indices.Create(
		"news",
		m.esClient.Indices.Create.WithContext(context.Background()),
	)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}
	fmt.Println("Create Index Response: ", createIndexRes)

	mapping := `
	{
	    "properties": {
	        "location": { "type": "geo_point" }
	    }
	}`
	_, err = m.esClient.Indices.PutMapping(
		[]string{"news"},
		strings.NewReader(mapping),
		m.esClient.Indices.PutMapping.WithContext(context.Background()),
	)
	if err != nil {
		log.Fatalf("Failed to put mapping: %v", err)
	}
}
