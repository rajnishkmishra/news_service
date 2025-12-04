package utils

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type Context struct {
	CorrelationID string
	Ctx           context.Context
}

func CreateContextWithTimeout(ginCtx *gin.Context, timeout time.Duration) (*Context, context.CancelFunc) {
	request := ginCtx.Request
	correlationID := getCorrelationID(request)
	ctx := &Context{
		CorrelationID: correlationID,
		Ctx:           ginCtx,
	}
	newCtx, cancelFunc := context.WithTimeout(ctx.Ctx, timeout)
	ctx.Ctx = newCtx
	return ctx, func() {
		cancelFunc()
	}
}

func getCorrelationID(request *http.Request) string {
	if request == nil || request.Header == nil {
		return getUniqID()
	}
	correlationID := request.Header.Get("CORRELATION-ID")
	if correlationID == "" {
		correlationID = getUniqID()
	}
	return correlationID
}

func getUniqID() string {
	randID := strings.Replace(uuid.Must(uuid.NewV4(), nil).String(), "-", "", -1)
	timeID := strings.Replace(uuid.Must(uuid.NewV1(), nil).String(), "-", "", -1)
	return timeID[:len(timeID)/2] + randID[len(randID)/2:]
}
