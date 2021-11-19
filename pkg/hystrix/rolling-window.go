package hystrix

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type RollingWindow struct {
	sync.RWMutex
	broken bool
	// 滑动窗口大小
	size int
	// 桶队列
	buckets []*Bucket
	// 触发熔断的请求总数阈值
	reqThreshold int
	// 出发熔断的失败率阈值
	failedThreshold float64
	// 上次熔断发生时间
	lastBreakTime time.Time
	seeker        bool
	// 熔断恢复的时间间隔
	brokeTimeGap time.Duration
}

// 新建滑动窗口
func NewRollingWindow(
	size int,
	reqThreshold int,
	failedThreshold float64,
	brokeTimeGap time.Duration,
) *RollingWindow {
	return &RollingWindow{
		size:            size,
		buckets:         make([]*Bucket, 0, size),
		reqThreshold:    reqThreshold,
		failedThreshold: failedThreshold,
		brokeTimeGap:    brokeTimeGap,
	}
}

// 追加一个新桶
func (r *RollingWindow) AppendBucket() {
	r.Lock()
	defer r.Unlock()
	r.buckets = append(r.buckets, NewBucket())
	if !(len(r.buckets) < r.size+1) {
		r.buckets = r.buckets[1:]
	}
}

// 获取当前队列末端的桶
func (r *RollingWindow) GetBucket() *Bucket {
	if len(r.buckets) == 0 {
		r.AppendBucket()
	}
	return r.buckets[len(r.buckets)-1]
}

// 在桶中记录当次结果
func (r *RollingWindow) RecordReqResult(result bool) {
	r.GetBucket().Record(result)
}

// 展示当前滑动窗口的所有桶状态
func (r *RollingWindow) ShowAllBucket() {
	for _, v := range r.buckets {
		fmt.Printf("id: [%v] | total: [%d] | failed: [%d]\n", v.Timestamp, v.Total, v.Failed)
	}
}

// 启动滑动窗口
func (r *RollingWindow) Launch() {
	go func() {
		for {
			r.AppendBucket()
			time.Sleep(time.Millisecond * 100)
		}
	}()
}

// 根据当前滑动窗口判断是否需要触发熔断
func (r *RollingWindow) BreakJudgement() bool {
	r.RLock()
	defer r.RUnlock()
	total := 0
	failed := 0
	for _, v := range r.buckets {
		total += v.Total
		failed += v.Failed
	}
	if float64(failed)/float64(total) > r.failedThreshold && total > r.reqThreshold {
		return true
	}
	return false
}

// 监控华东窗口的总失败次数与是否开启熔断
func (r *RollingWindow) Monitor() {
	go func() {
		for {
			if r.broken {
				if r.OverBrokenTimeGap() {
					r.Lock()
					r.broken = false
					r.Unlock()
				}
				continue
			}
			if r.BreakJudgement() {
				r.Lock()
				r.broken = true
				r.lastBreakTime = time.Now()
				r.Unlock()
			}
		}
	}()
}

// 查询是否超过熔断间隔期
func (r *RollingWindow) OverBrokenTimeGap() bool {
	return time.Since(r.lastBreakTime) > r.brokeTimeGap
}

// 每隔一秒展示当前是否处于熔断状态
func (r *RollingWindow) ShowStatus() {
	go func() {
		for {
			log.Println(r.broken)
			time.Sleep(time.Second)
		}
	}()
}

// 获取当前熔断状态
func (r *RollingWindow) Broken() bool {
	return r.broken
}

// 设置探测器状态
func (r *RollingWindow) SetSeeker(status bool) {
	r.Lock()
	defer r.Unlock()
}

// 获知探测是否被派出
func (r *RollingWindow) Seeker() bool {
	return r.seeker
}
