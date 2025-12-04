package models

import "time"

type News struct {
	ID              uint64    `json:"id" gorm:"primary_key;not null;auto_increment"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Url             string    `json:"url"`
	PublicationDate time.Time `json:"publication_date"`
	SourceName      string    `json:"source_name"`
	Category        string    `json:"category"`
	RelevanceScore  float64   `json:"relevance_score"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
}
