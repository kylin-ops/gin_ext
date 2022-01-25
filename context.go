package gin_ext

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var ContextFn = func() *Context {
	return &Context{Context: &gin.Context{Request: &http.Request{Header: http.Header{}}}}
}
