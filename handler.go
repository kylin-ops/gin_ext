package gin_ext

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kylin-ops/gin_ext/grequest"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	traceIdName  = "traceID"
	spanIdName   = "spanID"
	parentIdName = "parentID"
)

var LogDebug bool

func RandStr(l int) string {
	rand.Seed(time.Now().UnixNano())
	r := "abcdefghijklmnopqrstuvwxy0123456789"
	var d []byte
	for i := 0; i < l; i++ {
		d = append(d, r[rand.Intn(len(r))])
	}
	return string(d)
}

var (
	Logger  *logrus.Logger
	SLogger *logrus.Logger
	AppName string
	Host    string
	Ip      string
)

type Context struct {
	*gin.Context
}

type HandlerFunc func(*Context)

func HandlerExt(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := Context{}
		ctx.Context = c
		h(&ctx)
	}
}

func (c *Context) GetTraceID() string {
	return c.Request.Header.Get(traceIdName)
}

func (c *Context) GetSpanID() string {
	return c.Request.Header.Get(spanIdName)
}

func (c *Context) GetParentID() string {
	return c.Request.Header.Get(parentIdName)
}

func (c *Context) QueryInt(key string) (int, error) {
	val := c.Query(key)
	return strconv.Atoi(val)
}

//code 0代表正常，非0代表不正常
func (c *Context) ApiResponse(code int, data interface{}, msg interface{}) {
	responseData := gin.H{"code": fmt.Sprintf("%d", code), "info": data, "msg": msg}
	c.Set("responseData", responseData)
	c.JSON(http.StatusOK, responseData)
}

func (c *Context) Logger(logType, level, msg string, args ...interface{}) {
	jfields := logrus.Fields{
		traceIdName:  c.GetTraceID(),
		spanIdName:   c.GetSpanID(),
		parentIdName: c.GetParentID(),
		"type":       logType,
		"appName":    AppName,
		"host":       Host,
		"ip":         Ip,
	}
	sfields := logrus.Fields{
		"type": logType,
	}
	if logType == "ACCESS" {
		responseTime := time.Now().UnixNano()
		requestTime := c.GetInt64("requestTime")
		info := fmt.Sprintf("method=%s | responseTime=%f | clientIp=%s | code=%d | uri=%s | msg=",
			c.Request.Method, float64(responseTime-requestTime)/1e6, c.ClientIP(),
			c.Writer.Status(), c.Request.URL.String(),
		)
		msg = info + msg
	}
	jlogger := Logger.WithFields(jfields)
	slogger := SLogger.WithFields(sfields)
	switch level {
	case "info":
		jlogger.Infof(msg, args...)
		slogger.Infof(msg, args...)
	case "debug":
		jlogger.Debugf(msg, args...)
		slogger.Debugf(msg, args...)
	case "warn":
		jlogger.Warnf(msg, args...)
		slogger.Warnf(msg, args...)
	case "error":
		jlogger.Errorf(msg, args...)
		slogger.Errorf(msg, args...)
	case "tranc":
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		jlogger.Errorf(msg, args...)
		slogger.Errorf(msg, args...)
		//logger.Panicf(msg, args...)
		jlogger.Error(string(buf[:n]))
		slogger.Error(string(buf[:n]))
	}
}

func (c *Context) LogInfof(msg string, args ...interface{}) {
	c.Logger("EVENT", "info", msg, args...)
}

func (c *Context) LogDebugf(msg string, args ...interface{}) {
	c.Logger("EVENT", "debug", msg, args...)
}

func (c *Context) LogWarnf(msg string, args ...interface{}) {
	c.Logger("EVENT", "warn", msg, args...)
}

func (c *Context) LogErrorf(msg string, args ...interface{}) {
	c.Logger("EVENT", "error", msg, args...)
}

func (c *Context) LogTrancf(msg string, args ...interface{}) {
	c.Logger("EVENT", "Trancf", msg, args...)
}

func (c *Context) gRequest(url, method string, options ...*grequest.RequestOptions) (*grequest.Response, error) {
	method = strings.ToUpper(method)
	var option *grequest.RequestOptions
	if options == nil {
		option = &grequest.RequestOptions{Header: grequest.Header{}}
	} else {
		option = options[0]
		if option.Header == nil {
			option.Header = grequest.Header{}
		}
	}
	traceID := c.Request.Header.Get("traceID")
	if traceID == "" {
		traceID = uuid.New().String()
		c.Request.Header.Set("traceID", traceID)
	}
	spanID := c.Request.Header.Get("spanID")
	if spanID == "" {
		spanID = uuid.New().String()
		c.Request.Header.Set("spanID", spanID)
	}
	option.Header["traceId"] = traceID
	option.Header["spanId"] = spanID
	option.Header["app_name"] = AppName

	if LogDebug {
		bodyData, _ := json.Marshal(option.Data)
		paramData, _ := json.Marshal(option.Params)
		c.LogDebugf("请求method:%s | 请求url:%s | 请求参数:%s | 请求数据:%s", method, url, string(paramData), string(bodyData))
	}
	resp, err := grequest.Request(url, method, option)
	if LogDebug && resp != nil {
		respData, _ := ioutil.ReadAll(resp.Response.Body)
		paramData, _ := json.Marshal(option.Params)
		resp.Response.Body = ioutil.NopCloser(bytes.NewBuffer(respData))
		c.LogDebugf("请求method:%s | 请求url:%s | 请求参数:%s | 响应状态码:%d | 响应数据:%s", method, url, string(paramData), resp.StatusCode(), string(respData))
	}
	if LogDebug && err != nil {
		paramData, _ := json.Marshal(option.Params)
		c.LogDebugf("请求method:%s | 请求url:%s | 请求参数:%s | 错误信息:%s", method, url, string(paramData), err.Error())
	}

	if resp == nil {
		return nil, errors.New("响应response是空")
	}
	return resp, err
}

func (c *Context) GetRequest(url string, options ...*grequest.RequestOptions) (*grequest.Response, error) {
	return c.gRequest(url, "get", options...)
}

func (c *Context) PostRequest(url string, options ...*grequest.RequestOptions) (*grequest.Response, error) {
	return c.gRequest(url, "post", options...)
}

func (c *Context) DeleteRequest(url string, options ...*grequest.RequestOptions) (*grequest.Response, error) {
	return c.gRequest(url, "delete", options...)
}

func (c *Context) PutRequest(url string, options ...*grequest.RequestOptions) (*grequest.Response, error) {
	return c.gRequest(url, "put", options...)
}

func (c *Context) PatchRequest(url string, options ...*grequest.RequestOptions) (*grequest.Response, error) {
	return c.gRequest(url, "patch", options...)
}

func (c *Context) HeadRequest(url string, options ...*grequest.RequestOptions) (*grequest.Response, error) {
	return c.gRequest(url, "header", options...)
}

func (c *Context) DownloadRequest(url, dest string) error {
	return grequest.DownloadFile(url, dest)
}
