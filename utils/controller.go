package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ControllerOptions[RequestType any, ResponseType any] struct {
	name          string
	impl          func(*Context, RequestType) (ResponseType, WrapperError)
	requestBinder func(*gin.Context) (RequestType, error)
	timeout       time.Duration
}

type DebugInfo struct {
	CorrelationId string `json:"correlation-id"`
}

func getBinder[RequestType any](ginCtx *gin.Context) (RequestType, error) {
	var request RequestType
	err := ginCtx.ShouldBindUri(&request)
	if err != nil {
		return request, err
	}
	err = ginCtx.ShouldBindQuery(&request)
	if err != nil {
		return request, err
	}
	return request, err
}

func NewOptions[RequestType, ResponseType any](impl func(*Context, RequestType) (ResponseType, WrapperError)) ControllerOptions[RequestType, ResponseType] {
	return ControllerOptions[RequestType, ResponseType]{
		impl:          impl,
		requestBinder: getBinder[RequestType],
		timeout:       time.Duration(50) * time.Second,
	}
}

func (c ControllerOptions[RequestType, ResponseType]) ForPost() ControllerOptions[RequestType, ResponseType] {
	c.requestBinder = postBinder[RequestType]
	return c
}

func postBinder[RequestType any](ginCtx *gin.Context) (RequestType, error) {
	request, err := getBinder[RequestType](ginCtx)
	if err != nil {
		return request, err
	}

	err = ginCtx.ShouldBindJSON(&request)
	errString := fmt.Sprintf("%v", err)
	if errString == "EOF" {
		err = nil
	}
	return request, err
}

func executeController[RequestType any, ResponseType any](
	ctx *Context, request RequestType, options ControllerOptions[RequestType, ResponseType]) (ResponseType, WrapperError) {

	response, werr := options.impl(ctx, request)
	return response, werr
}

func Controller[RequestType any, ResponseType any](options ControllerOptions[RequestType, ResponseType]) func(ctx *gin.Context) {
	return func(ginCtx *gin.Context) {
		defer Recovery()
		ctx, cancelFunc := CreateContextWithTimeout(ginCtx, options.timeout)
		defer cancelFunc()

		request, err := options.requestBinder(ginCtx)
		if err != nil {
			ginCtx.SecureJSON(http.StatusBadRequest, gin.H{
				"error":      "Invalid request.",
				"debug_info": DebugInfo{CorrelationId: ctx.CorrelationID},
			})
			logrus.Error(err.Error())
			return
		}
		logrus.WithContext(ctx.Ctx).Info(fmt.Sprintf("request for %v Value : %+v", options.name, request))

		res, werr := executeController[RequestType, ResponseType](ctx, request, options)
		if werr != nil {
			logrus.Error(werr.Error())
			response := gin.H{
				"error":      fmt.Sprintf("%v", werr),
				"debug_info": DebugInfo{CorrelationId: ctx.CorrelationID},
			}
			ginCtx.SecureJSON(werr.HttpCode(), response)
		} else {
			ginCtx.SecureJSON(http.StatusOK, gin.H{"data": res, "debug_info": DebugInfo{CorrelationId: ctx.CorrelationID}})
		}
	}
}
