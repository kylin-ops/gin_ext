package middleware

import (
	"encoding/json"
	"github.com/kylin-ops/gin_exit"

	"github.com/google/uuid"

	//"gin_demo/internal/logger"
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	traceIdName  = "traceID"
	spanIdName   = "spanID"
	parentIdName = "parentID"
)

func AccessLoggerHandler(ctx *gin_ext.Context) {
	requestTime := time.Now().UnixNano()
	ctx.Set("requestTime", requestTime)
	traceId := ctx.Request.Header.Get(traceIdName)
	if traceId == "" {
		traceId = uuid.New().String()
		ctx.Request.Header.Set(traceIdName, traceId)
	}
	spanID := ctx.Request.Header.Get(spanIdName)
	if spanID != "" {
		ctx.Request.Header.Set(parentIdName, spanID)
	}
	ctx.Request.Header.Set(spanIdName, uuid.New().String())
	body, _ := ioutil.ReadAll(ctx.Request.Body)
	if len(body) > 0 {
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
	ctx.Logger("ACCESS", "info", "请求体:"+string(body))

	ctx.Next()
	code := ctx.Writer.Status()

	responseData, _ := ctx.Get("responseData")
	var d []byte
	if responseData != nil {
		d, _ = json.Marshal(responseData)
	}
	if code > 199 && code < 400 {
		ctx.Logger("ACCESS", "info", "响应体:"+string(d))
	} else {
		ctx.Logger("ACCESS", "error", "响应体:"+string(d))
	}
}

func RecoverHandler(ctx *gin_ext.Context) {
	defer func() {
		if r := recover(); r != nil {
			//Info(fmt.Sprintf("%v", r))
			ctx.String(http.StatusInternalServerError, "%v", r)
			ctx.Abort()
		}
	}()
	ctx.Next()
}
