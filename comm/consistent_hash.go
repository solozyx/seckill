package comm

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// 一致性哈希取值范围 [0,2^32-1]
type uint32slice []uint32

func (x uint32slice) Len() int {
	return len(x)
}

// 比对两个数大小
func (x uint32slice) Less(i, j int) bool {
	return x[i] < x[j]
}

// 切片中两个值交换
func (x uint32slice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

// 保存一致性hash信息
type Consistent struct {
	// hash环 key为哈希值 值为节点信息
	circleHashNodeMap map[uint32]string
	// 已经排序的节点hash切片
	sortedHashes uint32slice
	// 虚拟节点个数 用来增加hash的平衡性 避免数据倾斜问题
	virtualNodeCount int
	// map 读写锁
	sync.RWMutex
}

// 一致性哈希实例构造函数 设置默认节点数量
func NewConsistent() *Consistent {
	return &Consistent{
		// 初始化变量
		circleHashNodeMap: make(map[uint32]string),
		// 设置虚拟节点个数
		virtualNodeCount: 20,
	}
}

// 自动生成key
func (c *Consistent) generateKey(element string, index int) (key string) {
	// 副本 key 生成逻辑
	return element + strconv.Itoa(index)
}

// 获取hash位置
func (c *Consistent) hashKey(key string) (hash uint32) {
	if len(key) < 64 {
		// 声明一个字节数组长度为64
		var scratch [64]byte
		// 拷贝key数据到数组中
		copy(scratch[:], key)
		// 使用IEEE 多项式返回数据的CRC-32校验和 通过该函数计算哈希值
		return crc32.ChecksumIEEE(scratch[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

// 更新排序 方便查找
func (c *Consistent) updateSortedHashes() {
	hashes := c.sortedHashes[:0]
	// 判断切片容量 是否过大 如果过大则重置
	if cap(c.sortedHashes)/(c.virtualNodeCount*4) > len(c.circleHashNodeMap) {
		hashes = nil
	}
	// 添加 hash
	// for hashKey := range c.circleHashNodeMap {
	for hashKey, _ := range c.circleHashNodeMap {
		hashes = append(hashes, hashKey)
	}
	// 对所有节点hash值进行排序 方便之后进行二分查找
	sort.Sort(hashes)
	// 重新赋值
	c.sortedHashes = hashes
}

// 向hash环中添加1个节点
func (c *Consistent) Add(element string) {
	// 在并发场景 对map的写操作需要加锁
	// 加锁
	c.Lock()
	// 解锁
	defer c.Unlock()
	c.add(element)
}

func (c *Consistent) add(element string) {
	// 循环虚拟节点设置副本
	for i := 0; i < c.virtualNodeCount; i++ {
		// 把虚拟节点映射到hash环中
		c.circleHashNodeMap[c.hashKey(c.generateKey(element, i))] = element
	}
	// 更新排序
	c.updateSortedHashes()
}

// 向hash环中删除1个节点
func (c *Consistent) Remove(element string) {
	c.Lock()
	defer c.Unlock()
	c.remove(element)
}

func (c *Consistent) remove(element string) {
	for i := 0; i < c.virtualNodeCount; i++ {
		delete(c.circleHashNodeMap, c.hashKey(c.generateKey(element, i)))
	}
	c.updateSortedHashes()
}

// 根据数据标示获取最近的服务器节点信息
func (c *Consistent) Get(element string) (string, error) {
	// 在并发场景 对map的读操作需要加读锁
	// 加读锁
	c.RLock()
	// 解读锁
	defer c.RUnlock()

	if len(c.circleHashNodeMap) == 0 {
		return "", errors.New("error : hash circle has no data")
	}

	// 计算hash值
	key := c.hashKey(element)
	i := c.search(key)
	return c.circleHashNodeMap[c.sortedHashes[i]], nil
}

// 顺时针查找最近的服务端节点
func (c *Consistent) search(key uint32) int {
	// 查找算法
	f := func(x int) bool {
		return c.sortedHashes[x] > key
	}
	// 使用 二分查找 来搜索指定切片满足条件的最小值
	i := sort.Search(len(c.sortedHashes), f)
	// 如果超出范围则设置i=0
	if i >= len(c.sortedHashes) {
		i = 0
	}
	return i
}
