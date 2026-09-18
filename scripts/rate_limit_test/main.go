// 限流并发验证脚本
// 运行方式（在项目根目录执行）：
//   1. 启动服务： go run demo.go
//   2. 另开终端运行脚本： go run ./scripts/rate_limit_test/main.go
//
// 预期结果：
//   - 手机验证码 PeriodLimit（quota=10/60s）：20 个并发请求 → 成功 10 / 限流 10
//   - 创建订单 TokenLimiter（rate=100/s, burst=200）：300 个并发请求 → 成功≈200 / 限流≈100
//
// 注：项目当前对限流响应已统一为 HTTP 429 + JSON {"code":-1,"message":"请求过于频繁，请稍后再试"}
// 脚本只通过 HTTP 状态码区分成功/限流，样本行直接展示原始响应体，不做任何解析。
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultBaseURL = "http://127.0.0.1:8887"
	smsPath        = "/api/v1/sms/send"
	orderPath      = "/api/v1/orders"

	smsQuota    = 10  // SendSmsQuota: 每分钟 10 次
	orderBurst  = 200 // OrdersBurst
	orderRate   = 100 // OrdersRate: 每秒 100
	smsTotal    = 20  // 验证码并发请求总数
	orderTotal  = 300 // 订单并发请求总数
	sampleLimit = 4   // 每类样本最多展示条数，避免刷屏
)

var client = &http.Client{Timeout: 10 * time.Second}

// stats 统计并发测试结果，并收集少量样本（线程安全）
type stats struct {
	success int32
	limited int32
	other   int32
	mu      sync.Mutex
	succS   []string
	limS    []string
	othS    []string
}

func (s *stats) addSuccess(idx int, msg string) {
	atomic.AddInt32(&s.success, 1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.succS) < sampleLimit {
		s.succS = append(s.succS, fmt.Sprintf("  [#%02d] 成功: %s", idx, msg))
	}
}

func (s *stats) addLimited(idx int, msg string) {
	atomic.AddInt32(&s.limited, 1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.limS) < sampleLimit {
		s.limS = append(s.limS, fmt.Sprintf("  [#%02d] 限流: %s", idx, msg))
	}
}

func (s *stats) addOther(idx int, msg string) {
	atomic.AddInt32(&s.other, 1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.othS) < sampleLimit {
		s.othS = append(s.othS, fmt.Sprintf("  [#%02d] 其他: %s", idx, msg))
	}
}

// classify 响应分类：HTTP 200 成功；HTTP 429 限流；其他异常
// 样本直接展示原始响应体，不做任何解析
func classify(resp *http.Response, body []byte, idx int, s *stats) {
	raw := fmt.Sprintf("HTTP %d, body=%s", resp.StatusCode, string(body))
	switch resp.StatusCode {
	case http.StatusOK:
		s.addSuccess(idx, raw)
	case http.StatusTooManyRequests:
		s.addLimited(idx, raw)
	default:
		s.addOther(idx, raw)
	}
}

func main() {
	base := flag.String("base", defaultBaseURL, "目标服务地址")
	flag.Parse()

	fmt.Println("================ 限流并发验证脚本 ================")
	fmt.Printf("目标服务: %s\n", *base)
	fmt.Printf("配置:\n  - 手机验证码 PeriodLimit: quota=%d / 60s\n", smsQuota)
	fmt.Printf("  - 创建订单 TokenLimiter:  rate=%d/s, burst=%d\n", orderRate, orderBurst)
	fmt.Println()

	if !ping(*base) {
		fmt.Printf("服务无法访问，请确认 demo 服务已启动: %s\n", *base)
		return
	}
	fmt.Printf("服务可达，开始测试...\n\n")

	testSendSms(*base)
	fmt.Println()
	testCreateOrder(*base)

	fmt.Printf("\n================ 验证结束 ================\n")
}

func ping(base string) bool {
	resp, err := client.Get(base + smsPath)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// testSendSms 测试固定窗口配额限流
// 配置 quota=10/60s，同一手机号并发 20 个请求
func testSendSms(base string) {
	const phone = "13800138000"
	fmt.Println("---------------- 测试 1: 手机号验证码（固定窗口 PeriodLimit） ----------------")
	fmt.Printf("同一手机号 %s 并发 %d 个请求，配置 quota=%d/60s\n", phone, smsTotal, smsQuota)

	st := &stats{}
	concurrentRun(base+smsPath, smsTotal, func(idx int) []byte {
		body, _ := json.Marshal(map[string]string{"phone": phone})
		return body
	}, classify, st)

	printStats(st, smsTotal, fmt.Sprintf("预期: 成功=%d, 限流=%d -> 配额被切到固定窗口", smsQuota, smsTotal-smsQuota))
}

// testCreateOrder 测试令牌桶限流
// 配置 rate=100/s, burst=200，并发 300 个请求
func testCreateOrder(base string) {
	fmt.Println("---------------- 测试 2: 创建订单（令牌桶 TokenLimiter） ----------------")
	fmt.Printf("并发 %d 个请求，配置 rate=%d/s, burst=%d\n", orderTotal, orderRate, orderBurst)

	st := &stats{}
	concurrentRun(base+orderPath, orderTotal, func(idx int) []byte {
		return []byte(fmt.Sprintf(`{"userId":"u%d","productId":"p1","quantity":1}`, idx))
	}, classify, st)

	printStats(st, orderTotal, fmt.Sprintf("预期: 成功≈%d, 限流≈%d", orderBurst, orderTotal-orderBurst))
	fmt.Println("注: 令牌补充率 100/s，请求若耗时较长则实际成功数会略高于 burst")
}

// concurrentRun 并发执行 total 个 HTTP POST 请求
//   - url: 完整请求地址
//   - total: 并发总数
//   - bodyFn: 根据 idx 生成请求体
//   - classify: 响应分类回调
//   - s: 样本与计数收集器
func concurrentRun(url string, total int, bodyFn func(idx int) []byte,
	classify func(resp *http.Response, body []byte, idx int, s *stats), s *stats) {
	var wg sync.WaitGroup
	gate := make(chan struct{}) // 同步门，让所有 goroutine 同时触发

	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-gate

			req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyFn(idx)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				s.addOther(idx, fmt.Sprintf("请求错误: %v", err))
				return
			}
			defer resp.Body.Close()
			rb, _ := io.ReadAll(resp.Body)

			classify(resp, rb, idx, s)
		}(i)
	}

	close(gate) // 同时放开所有 goroutine，制造并发
	wg.Wait()
}

// printStats 打印汇总结果和样本
func printStats(s *stats, total int, expected string) {
	fmt.Printf("结果: 成功=%d, 限流=%d, 其他=%d / 总计 %d\n",
		atomic.LoadInt32(&s.success), atomic.LoadInt32(&s.limited),
		atomic.LoadInt32(&s.other), total)
	for _, line := range s.succS {
		fmt.Println(line)
	}
	for _, line := range s.limS {
		fmt.Println(line)
	}
	for _, line := range s.othS {
		fmt.Println(line)
	}
	fmt.Println(expected)
}
