package llm_service

import (
	"encoding/json"
	"fmt"
	"log"
	"news_service/utils"

	"github.com/sirupsen/logrus"
	"google.golang.org/genai"
)

const (
	GEMINI_API_KEY = "GEMINI_API_KEY" //Put Gemini API key here
)

type LlmService struct{}

type LlmOutput struct {
	Entities []string `json:"entities"`
	Intent   []string `json:"intent"`
}

func NewLlmService() *LlmService {
	return &LlmService{}
}

func (l *LlmService) AnalyzeQuery(ctx *utils.Context, query string) (out *LlmOutput, err error) {

	prompt := fmt.Sprintf(`
	You extract news query info.

	Extract:
	- entities (people, orgs, locations)
	- intent ("latest", "category", "source", "nearby")

	Respond ONLY in JSON with no explanation, no markdown, no code block, no backticks.
	Example: {"entities": ["News18"], "intent": ["category"]}
	User Query: "%s"`, query)

	client, err := genai.NewClient(ctx.Ctx, &genai.ClientConfig{
		APIKey:  GEMINI_API_KEY,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Models.GenerateContent(
		ctx.Ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		logrus.WithContext(ctx.Ctx).Error(err)
		return
	}

	fmt.Println(result.Text())

	if err := json.Unmarshal([]byte(result.Text()), &out); err != nil {
		logrus.WithContext(ctx.Ctx).Error(
			fmt.Errorf("JSON parse error: %v\nRaw: %s", err, result.Text()))
		return nil, err
	}

	return
}

func (l *LlmService) GenerateSummary(ctx *utils.Context, articles []string) (summary []string, err error) {
	prompt := fmt.Sprintf(`
	Generate summary for given articles.
	And give output as an array of string as a JSON with no explanation, no markdown, no code block, no backticks.
	Also, use escape character if required.
	Articles: "%v"`, articles)

	client, err := genai.NewClient(ctx.Ctx, &genai.ClientConfig{
		APIKey:  GEMINI_API_KEY,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		logrus.WithContext(ctx.Ctx).Error(err)
		return
	}

	result, err := client.Models.GenerateContent(
		ctx.Ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		logrus.WithContext(ctx.Ctx).Error(err)
		return
	}

	fmt.Println(result.Text())

	summary = make([]string, 0)
	if err := json.Unmarshal([]byte(result.Text()), &summary); err != nil {
		logrus.WithContext(ctx.Ctx).Error(
			fmt.Errorf("JSON parse error: %v\nRaw: %s", err, result.Text()))
		return nil, err
	}

	return
}
