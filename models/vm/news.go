package vm

import "time"

type NewsResponse struct {
	Articles []News `json:"articles"`
	MetaResponse
}

type News struct {
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Url             string    `json:"url"`
	PublicationDate time.Time `json:"publication_date"`
	SourceName      string    `json:"source_name"`
	Category        []string  `json:"category"`
	RelevanceScore  float64   `json:"relevance_score"`
	LLMSummary      string    `json:"llm_summary"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
}

type NewsRaw struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Url             string   `json:"url"`
	PublicationDate string   `json:"publication_date"`
	SourceName      string   `json:"source_name"`
	Category        []string `json:"category"`
	RelevanceScore  float64  `json:"relevance_score"`
	Latitude        float64  `json:"latitude"`
	Longitude       float64  `json:"longitude"`
}

type NewsElastic struct {
	ID              uint64    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Url             string    `json:"url"`
	PublicationDate time.Time `json:"publication_date"`
	SourceName      string    `json:"source_name"`
	Category        []string  `json:"category"`
	RelevanceScore  float64   `json:"relevance_score"`
	Location        LatLon    `json:"location"`
}

type LatLon struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type FetchNewsRequest struct {
	Category string  `uri:"category" json:"category,omitempty"`
	Score    float64 `uri:"score" json:"score,omitempty"`
	Source   string  `uri:"source" json:"source,omitempty"`
	Lat      float64 `form:"lat" json:"lat,omitempty"`
	Long     float64 `form:"long" json:"long,omitempty"`
	Radius   int     `form:"radius" json:"radius,omitempty"`
	Query    string  `form:"q" json:"query,omitempty"`
	PaginationRequest
}

type PaginationRequest struct {
	PageNumber int64 `form:"p" json:"page_number,omitempty"`
	Limit      int64 `form:"l" json:"limit,omitempty"`
}

func NewPaginationRequest(p int64, l int64) PaginationRequest {
	return PaginationRequest{
		PageNumber: p,
		Limit:      l,
	}
}

func (p PaginationRequest) GetPageNumber() int64 {
	if p.PageNumber <= 0 {
		return 1
	}
	return p.PageNumber
}

func (p PaginationRequest) GetLimit() int64 {
	if p.Limit <= 0 {
		return 5
	}
	return p.Limit
}

type MetaResponse struct {
	TotalPages  int64                  `json:"total_pages"`
	TotalRecord int64                  `json:"total_records"`
	PageNumber  int64                  `json:"page_number"`
	SearchQuery map[string]interface{} `json:"search_query"`
}
