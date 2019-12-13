/**
分布式验证支持水平扩展 一致性哈希算法保证数据均匀存储和快速查找
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/solozyx/seckill/comm"
	"github.com/solozyx/seckill/conf"
	"github.com/solozyx/seckill/datasource"
	"github.com/solozyx/seckill/limit"
	"github.com/solozyx/seckill/model"
)

var (
	// 秒杀用户访问控制
	// 负载均衡器设置使用 [一致性哈希算法]  水平扩展,秒杀用户分布式鉴权
	consistentHash *comm.Consistent
	// 设置集群节点地址 使用内网IP访问快 公网IP访问慢
	hostArray = []string{conf.ServiceValidateIP1, conf.ServiceValidateIP2}
	// accessControl = &AccessControl{sources:make(map[int]interface{})}
	accessControl *limit.AccessControl
	// 本机IP
	localHost = "" // "127.0.0.1"
	rabbitmq  *datasource.RabbitMQ
)

func main() {
	// TODO:NOTICE 引入一致性哈希算法
	consistentHash = comm.NewConsistent()
	// 采用一致性hash算法 添加服务端节点
	for _, v := range hostArray {
		consistentHash.Add(v)
	}
	// 获取本机IP 192.168.174.134
	localIp, err := comm.GetLocalInterfaceIP()
	if err != nil {
		fmt.Println(err)
	}
	localHost = localIp
	// 限流 黑名单
	accessControl = limit.NewAccessControl()
	// rabbitmq
	rabbitmq = datasource.NewRabbitMQSimple(conf.SeckillQueueName)
	defer rabbitmq.Destroy()
	// TODO:WARNING 生产部署注意静态资源目录 前端通过ajax请求后端秒杀接口
	// 设置静态文件目录
	http.Handle("/html/",
		http.StripPrefix("/html/", http.FileServer(http.Dir("./fronted/web/static_product"))))
	// 设置资源目录
	http.Handle("/public/",
		http.StripPrefix("/public/", http.FileServer(http.Dir("./fronted/web/public"))))

	// 1.秒杀请求拦截器
	filter := comm.NewFilter()
	// TODO 优化注册拦截器中间件
	filter.RegisterFilterUri("/check_right", auth)
	filter.RegisterFilterUri("/check", auth)

	// 2.业务逻辑
	http.HandleFunc("/check_right", filter.Handle(checkRight))
	http.HandleFunc("/check", filter.Handle(check))

	// 3.启动服务
	http.ListenAndServe(fmt.Sprintf(":%s", conf.ServiceValidatePort), nil)
}

// 用户是否登录是否合法验证中间件
func auth(w http.ResponseWriter, r *http.Request) error {
	// 基于cookie的权限验证
	uidCookie, err := r.Cookie(conf.CookieName)
	if err != nil {
		return errors.New("cookie 用户 uid 获取失败！")
	}
	signCookie, err := r.Cookie(conf.CookieSign)
	if err != nil {
		return errors.New("cookie 用户加密串 sign 获取失败！")
	}
	// 对信息进行解密
	signByte, err := comm.Base64DecodeAesDecrypt(signCookie.Value)
	if err != nil {
		return errors.New("cookie 用户加密串 sign 已被篡改！")
	}
	if uidCookie.Value == string(signByte) {
		return nil
	}
	return errors.New("用户身份校验失败！")
}

// 分布式权限验证
func checkRight(w http.ResponseWriter, r *http.Request) {
	right := getDistributedRight(r)
	if !right {
		w.Write([]byte("false"))
		return
	}
	w.Write([]byte("true"))
	return
}

// 秒杀业务逻辑
func check(w http.ResponseWriter, r *http.Request) {
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(queryForm["productID"]) <= 0 || len(queryForm["productID"][0]) <= 0 {
		w.Write([]byte("false"))
		return
	}
	productString := queryForm["productID"][0]
	// 获取用户 cookie
	userCookie, err := r.Cookie(conf.CookieName)
	if err != nil {
		w.Write([]byte("false"))
		return
	}

	// 1.分布式权限验证
	right := getDistributedRight(r)
	if right == false {
		w.Write([]byte("false"))
		return
	}

	// 2.获取数量控制权限 防止秒杀出现超卖现象
	hostUrl := "http://" + conf.ServiceGetOneIP + ":" + conf.ServiceGetOnePort + "/getone"
	validateResp, validateBody, err := getCurl(hostUrl, r)
	if err != nil {
		w.Write([]byte("false"))
		return
	}

	if validateResp.StatusCode == 200 {
		if string(validateBody) == "true" {
			// 整合下单
			// 1.获取商品ID
			productID, err := strconv.ParseInt(productString, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 2.获取用户ID
			userID, err := strconv.ParseInt(userCookie.Value, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 3.消息队列异步下单
			message := model.NewMessage(userID, productID)
			byteMessage, err := json.Marshal(message)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			err = rabbitmq.PublishSimple(string(byteMessage))
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 成功抢购到商品 把用户设置黑名单
			accessControl.SetBlackListById(int(userID))
			w.Write([]byte("true"))
			return
		}
	}
	w.Write([]byte("false"))
	return
}

// TODO:NOTICE 获取分布式权限 入参用户请求
func getDistributedRight(r *http.Request) bool {
	// 获取用户UID
	uid, err := r.Cookie(conf.CookieName)
	if err != nil {
		return false
	}
	// 采用一致性hash算法 根据用户ID 判断获取具体机器
	host, err := consistentHash.Get(uid.Value)
	if err != nil {
		return false
	}

	if host == localHost {
		// 执行本机数据读取和校验
		return accessControl.GetDataFromLocal(uid.Value)
	} else {
		// 不是本机 充当代理 访问数据返回结果
		return getDataFromRemote(host, r)
	}
}

// 模拟请求
func getCurl(hostUrl string, r *http.Request) (response *http.Response, body []byte, err error) {
	uid, err := r.Cookie(conf.CookieName)
	if err != nil {
		return
	}
	sign, err := r.Cookie(conf.CookieSign)
	if err != nil {
		return
	}

	// 模拟接口访问
	client := &http.Client{}
	req, err := http.NewRequest("GET", hostUrl, nil)
	if err != nil {
		return
	}
	// 添加cookie到模拟的请求中 手动指定 排查多余 cookies
	cookieUid := &http.Cookie{Name: "uid", Value: uid.Value, Path: "/"}
	req.AddCookie(cookieUid)
	cookieSign := &http.Cookie{Name: "sign", Value: sign.Value, Path: "/"}
	req.AddCookie(cookieSign)
	// 获取返回结果
	response, err = client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	return
}

// 获取其它远程节点map
func getDataFromRemote(host string, r *http.Request) bool {
	hostUrl := "http://" + host + ":" + conf.ServiceGetOnePort + "/check_right"
	response, body, err := getCurl(hostUrl, r)
	if err != nil {
		return false
	}
	//判断状态
	if response.StatusCode == 200 {
		if string(body) == "true" {
			return true
		} else {
			return false
		}
	}
	return false
}
