package comm

import (
	"net/http"
	"strings"
)

// 声明1个新的数据类型 函数类型 Go可以把函数当作类型
type FilterHandle func(w http.ResponseWriter, r *http.Request) error
type WebHandle func(w http.ResponseWriter, r *http.Request)

// 秒杀用户请求拦截器
type Filter struct {
	// 用来存储需要拦截的URI
	filterMap map[string]FilterHandle
}

// Filter构造函数
func NewFilter() *Filter {
	return &Filter{filterMap: make(map[string]FilterHandle)}
}

// 注册拦截器
func (f *Filter) RegisterFilterUri(uri string, handler FilterHandle) {
	f.filterMap[uri] = handler
}

// 根据Uri获取对应 handle
func (f *Filter) GetFilterHandle(uri string) FilterHandle {
	return f.filterMap[uri]
}

// 执行拦截器，返回函数类型
func (f *Filter) Handle(webHandle WebHandle) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for path, handle := range f.filterMap {
			// if path == r.RequestURI {
			if strings.Contains(r.RequestURI, path) {
				// 执行拦截业务逻辑
				err := handle(w, r)
				if err != nil {
					w.Write([]byte(err.Error()))
					return
				}
				// 跳出循环
				break
			}
		}
		// 执行真正的web请求处理函数
		webHandle(w, r)
	}
}
