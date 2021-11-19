package hystrix

import (
	"sync"
	"time"
)

type Bucket struct {
	sync.RWMutex
	// 请求总数
	Total int
	// 失败总数
	Failed    int
	Timestamp time.Time
}

func NewBucket() *Bucket {
	return &Bucket{
		Timestamp: time.Now(),
	}
}

// 记录请求结果
func (b *Bucket) Record(result bool) {
	b.Lock()
	defer b.Unlock()
	if !result {
		b.Failed++
	}
	b.Total++
}
