package context

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/LyricTian/gin-admin/src/logger"
	"github.com/LyricTian/gin-admin/src/util"
	"github.com/gin-gonic/gin"
	"github.com/go-session/gin-session"
	"github.com/go-session/session"
	"github.com/pkg/errors"
)

// WrapContext 包装上下文
func WrapContext(ctx func(*Context), memo ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(memo) > 0 {
			c.Set(util.ContextKeyURLMemo, memo[0])
		}
		ctx(&Context{c})
	}
}

// NewContext 创建上下文实例
func NewContext(c *gin.Context) *Context {
	return &Context{c}
}

// Context 定义上下文
type Context struct {
	gctx *gin.Context
}

// GinContext 获取Gin上下文
func (a *Context) GinContext() *gin.Context {
	return a.gctx
}

// NewContext 创建上下文实例
func (a *Context) NewContext() context.Context {
	parent := context.Background()

	if v := a.GetTraceID(); v != "" {
		parent = util.NewTraceIDContext(parent, a.GetTraceID())
	}

	if v := a.GetUserID(); v != "" {
		parent = util.NewUserIDContext(parent, a.GetUserID())
	}

	return parent
}

// SessionStore 获取会话存储
func (a *Context) SessionStore() session.Store {
	return ginsession.FromContext(a.gctx)
}

// RefreshSession 更新会话
func (a *Context) RefreshSession() (session.Store, error) {
	return ginsession.Refresh(a.gctx)
}

// DestroySession 销毁会话
func (a *Context) DestroySession() error {
	return ginsession.Destroy(a.gctx)
}

// Param 获取路径参数(/foo/:id)
func (a *Context) Param(key string) string {
	return a.gctx.Param(key)
}

// Query 获取查询参数(/foo?id=)
func (a *Context) Query(key string) string {
	return a.gctx.Query(key)
}

// GetPageIndex 获取分页的页索引
func (a *Context) GetPageIndex() uint {
	if v := a.Query("current"); v != "" {
		if iv := util.S(v).Uint(); iv > 0 {
			return iv
		}
	}
	return 1
}

// GetPageSize 获取分页的页大小(最大50)
func (a *Context) GetPageSize() uint {
	if v := a.Query("pageSize"); v != "" {
		if iv := util.S(v).Uint(); iv > 0 {
			if iv > 50 {
				iv = 50
			}
			return iv
		}
	}
	return 10
}

// GetTraceID 获取追踪ID
func (a *Context) GetTraceID() string {
	return a.gctx.GetString(util.ContextKeyTraceID)
}

// GetUserID 获取用户ID
func (a *Context) GetUserID() string {
	return a.gctx.GetString(util.ContextKeyUserID)
}

// SetUserID 设定用户ID
func (a *Context) SetUserID(userID string) {
	a.gctx.Set(util.ContextKeyUserID, userID)
}

// ParseJSON 解析请求JSON
func (a *Context) ParseJSON(obj interface{}) error {
	if err := a.gctx.ShouldBindJSON(obj); err != nil {
		return errors.Wrap(err, "解析请求参数发生错误")
	}
	return nil
}

// ResBadRequest 响应客户端请求错误
func (a *Context) ResBadRequest(err error, code ...int) {
	a.ResError(err, http.StatusBadRequest, code...)
}

// ResInternalServerError 响应服务器错误
func (a *Context) ResInternalServerError(err error, code ...int) {
	status := http.StatusInternalServerError

	switch err {
	case util.ErrNotFound:
		status = http.StatusNotFound
	}

	a.ResError(err, status, code...)
}

// ResError 响应错误
func (a *Context) ResError(err error, status int, code ...int) {
	var item HTTPErrorItem

	if err != nil {
		ss := strings.Split(err.Error(), ": ")
		if len(ss) > 0 {
			item.Message = ss[0]
		}
	}

	if status >= 400 && status < 500 {
		if item.Message == "" {
			item.Message = "请求发生错误"
		}

		if err != nil {
			logger.System(a.NewContext()).
				WithField("error", err.Error()).
				Warnf("[请求错误] %s", item.Message)
		}
	} else if status >= 500 && status < 600 {
		if item.Message == "" {
			item.Message = "服务器发生错误"
		}

		if err != nil {
			type stackTracer interface {
				StackTrace() errors.StackTrace
			}

			entry := logger.System(a.NewContext())
			if stack, ok := err.(stackTracer); ok {
				entry = entry.WithField("error", fmt.Sprintf("%+v", stack.StackTrace()[:2]))
			}
			entry.Errorf("[服务器错误] %s", err.Error())
		}
	}

	if len(code) > 0 {
		item.Code = code[0]
	}
	a.ResJSON(status, HTTPError{Error: item})
}

// ResPage 响应分页数据
func (a *Context) ResPage(total int64, v interface{}) {
	a.ResSuccess(HTTPList{
		List: v,
		Pagination: &HTTPPagination{
			Total:    total,
			Current:  a.GetPageIndex(),
			PageSize: a.GetPageSize(),
		},
	})
}

// ResList 响应列表数据
func (a *Context) ResList(v interface{}) {
	a.ResSuccess(HTTPList{List: v})
}

// ResOK 响应OK
func (a *Context) ResOK() {
	a.ResSuccess(HTTPStatus{Status: StatusOK})
}

// ResSuccess 响应成功
func (a *Context) ResSuccess(v interface{}) {
	if v == nil {
		v = gin.H{}
	}
	a.ResJSON(http.StatusOK, v)
}

// ResJSON 响应JSON数据
func (a *Context) ResJSON(status int, v interface{}) {
	buf, err := util.JSONMarshal(v)
	if err != nil {
		a.ResInternalServerError(errors.Wrap(err, "JSON序列化发生错误"))
		return
	}
	a.gctx.Set(util.ContextKeyResBody, buf)
	a.gctx.Data(status, "application/json; charset=utf-8", buf)
	a.gctx.Abort()
}
