# 限流 Demo

基于 [go-zero](https://github.com/zeromicro/go-zero) 的限流示例项目，演示两种典型限流算法的实际应用：

- **固定窗口配额 `PeriodLimit`**：用于手机验证码接口，每手机号每分钟最多 10 次。
- **令牌桶 `TokenLimiter`**：用于创建订单接口，长期均值 100 QPS、瞬时峰值 200。

## 技术栈

- Go 1.26+
- go-zero v1.10.3
- goctl 1.10.2（代码脚手架）
- Redis（限流状态存储）

## 目录结构

```
.
├── demo.go                       # 程序入口
├── demo.api                       # API 定义文件
├── etc/demo.yaml                  # 配置文件
├── response/                     # 统一响应封装
├── utils/                        # 工具函数（随机数生成等）
├── internal/
│   ├── config/                   # 配置定义
│   ├── svc/                      # ServiceContext，装配限流中间件
│   ├── handler/                  # HTTP 路由与 handler
│   ├── logic/                    # 业务逻辑
│   │   ├── sendSms/              # 验证码发送逻辑
│   │   └── order/                # 订单创建逻辑
│   ├── middleware/               # 中间件
│   │   ├── requestlogmiddleware.go      # 全局请求日志
│   │   ├── sendsmslimitmiddleware.go   # PeriodLimit 限流
│   │   └── orderlimitmiddleware.go     # TokenLimiter 限流
│   └── types/                    # 请求/响应类型
└── scripts/rate_limit_test/      # 限流并发验证脚本
```

## API 接口

| 接口 | 方法 | 路由 | 限流算法 | 限流参数 |
| --- | --- | --- | --- | --- |
| 发送验证码 | POST | `/api/v1/sms/send` | PeriodLimit | 10 次/60s（按手机号） |
| 创建订单 | POST | `/api/v1/orders` | TokenLimiter | rate=100/s, burst=200 |

### 请求 / 响应示例

```bash
# 发送验证码
curl -X POST http://127.0.0.1:8887/api/v1/sms/send \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000"}'
# {"code":-1,"message":"OK","data":{"code":"831204"}}

# 创建订单
curl -X POST http://127.0.0.1:8887/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"userId":"u1","productId":"p1","quantity":1}'
# {"code":-1,"message":"OK","data":{"orderId":"...","message":"支付成功"}}
```

被限流时统一返回 HTTP 429 + `{"code":-1,"message":"请求过于频繁，请稍后再试"}`。

## 快速开始

### 1. 准备依赖

确保本地已启动 Redis（默认 `127.0.0.1:6379`）。

### 2. 启动服务

```bash
go run demo.go
```

启动后监听 `http://127.0.0.1:8887`。

### 3. 运行限流验证脚本

另开一个终端执行：

```bash
go run ./scripts/rate_limit_test/main.go
```

预期结果：

| 测试 | 接口 | 并发数 | 预期成功 | 预期限流 |
| --- | --- | --- | --- | --- |
| 验证码（PeriodLimit） | `/sms/send` | 20（同手机号） | 10 | 10 |
| 订单（TokenLimiter） | `/orders` | 300 | ≈200 | ≈100 |

## 配置说明

[etc/demo.yaml](etc/demo.yaml)：

```yaml
RateLimit:
  OrdersRate: 100   # 每秒允许的创建订单数量
  OrdersBurst: 200  # 允许的最大创建订单数量（桶容量）
  SendSmsQuota: 10  # 手机接收验证码每分钟次数
```

## 参考

- [go-zero 官方文档](https://go-zero.dev/)
