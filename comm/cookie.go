package comm

import (
	"net/http"

	"github.com/kataras/iris"
)

// 设置全局 cookie
func SetGlobalCookie(ctx iris.Context, name string, value string) {
	ctx.SetCookie(&http.Cookie{Name: name, Value: value, Path: "/"})
}
