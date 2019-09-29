/**
秒杀商品数量控制
*/
package main

import (
	"log"
	"net/http"
	"sync"
)

var (
	sum int64 = 0
	// 预存秒杀商品数量 秒杀过程中不能超过此数量
	productNum int64 = 1000 * 1000
	// 计数
	count int64 = 0
	// 互斥锁保护 sum productNum count 声明完成后是类型零值 可直接使用
	mutex sync.Mutex
)

func main() {
	http.HandleFunc("/getone", getProduct)
	err := http.ListenAndServe(":8084", nil)
	if err != nil {
		log.Fatal("Err:", err)
	}
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	if getOneProduct() {
		w.Write([]byte("true"))
		return
	}
	w.Write([]byte("false"))
	return
}

// 获取秒杀商品
func getOneProduct() bool {
	// 高并发多协程访问普通变量
	// 加互斥锁
	mutex.Lock()
	// 解互斥锁
	defer mutex.Unlock()

	count += 1
	// 判断数据是否超限
	if count%100 == 0 {
		if sum < productNum {
			sum += 1
			// 成功抢购1个商品
			return true
		}
	}
	return false
}
