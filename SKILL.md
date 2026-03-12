---
name: go-redis-skill
description: Redis官方Go客户端库(go-redis/v9)开发指南。当用户在Go项目中需要集成Redis、配置连接池、实现缓存、分布式锁、限流、消息队列、Pub/Sub、Pipeline、事务、Redis Cluster集群、Sentinel哨兵、Streams流、Lua脚本、OpenTelemetry监控等功能时使用。支持连接配置优化、性能调优、故障排查、测试mock等完整开发场景。22k stars，BSD-2-Clause许可。
triggers:
  - Go项目需要集成Redis
  - 使用go-redis库
  - 配置Redis连接池/集群/Sentinel
  - 实现Redis缓存/分布式锁/限流
  - Redis性能调优和问题排查
---

# go-redis 开发指南

> **官方 Redis Go 客户端** - 22k stars，类型安全，功能完整
> **版本**: v9.x (推荐) | **许可证**: BSD-2-Clause

## 快速参考

### 安装

```bash
go get github.com/redis/go-redis/v9
```

### 最小示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/redis/go-redis/v9"
)

func main() {
    ctx := context.Background()

    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    defer rdb.Close()

    // 测试连接
    if err := rdb.Ping(ctx).Err(); err != nil {
        log.Fatal(err)
    }

    // 基础操作
    rdb.Set(ctx, "key", "value", 0)
    val, err := rdb.Get(ctx, "key").Result()
    if err == redis.Nil {
        fmt.Println("key 不存在")
    }
}
```

## 配置选项

### 基础配置

```go
rdb := redis.NewClient(&redis.Options{
    // 连接地址
    Addr: "localhost:6379",

    // 认证
    Username: "user",           // Redis 6.0+ ACL 用户
    Password: "password",       // 密码
    DB:       0,                // 数据库编号

    // 协议版本
    Protocol: 3,                // 3=RESP3, 2=RESP2
})
```

### 连接池配置（生产推荐）

```go
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",

    // 连接池
    PoolSize:     100,              // 最大连接数（推荐：CPU核心数 * 10）
    PoolTimeout:  4 * time.Second,  // 连接池超时
    MinIdleConns: 10,               // 最小空闲连接

    // 超时设置
    DialTimeout:  5 * time.Second,  // 连接超时
    ReadTimeout:  3 * time.Second,  // 读超时
    WriteTimeout: 3 * time.Second,  // 写超时

    // 缓冲区（v9.12+ 默认 32KiB）
    ReadBufferSize:  32 * 1024,     // 读缓冲区
    WriteBufferSize: 32 * 1024,     // 写缓冲区

    // 连接保活
    ConnMaxLifetime: 30 * time.Minute, // 连接最大生命周期
    ConnMaxIdleTime: 5 * time.Minute,  // 空闲连接超时
})
```

### TLS 安全连接

```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Username: "user",
    Password: "password",
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
        InsecureSkipVerify: false, // 生产环境必须验证证书
    },
})

// 使用 URL 连接
opts, _ := redis.ParseURL("rediss://user:password@localhost:6379/0")
rdb := redis.NewClient(opts)
```

## 核心功能

### 字符串操作

```go
// SET
err := rdb.Set(ctx, "key", "value", 0).Err()
err := rdb.Set(ctx, "key", "value", 10*time.Second).Err() // 带过期

// SETNX（不存在才设置）
ok, err := rdb.SetNX(ctx, "key", "value", 10*time.Second).Result()

// GET
val, err := rdb.Get(ctx, "key").Result()

// MSET/MGET（批量操作）
rdb.MSet(ctx, "key1", "val1", "key2", "val2")
vals, err := rdb.MGet(ctx, "key1", "key2").Result()

// INCR/DECR（计数器）
n, err := rdb.Incr(ctx, "counter").Result()
```

### Hash 操作

```go
// HSET
rdb.HSet(ctx, "user:1", "name", "张三", "age", 25)

// HGET/HGETALL
val, _ := rdb.HGet(ctx, "user:1", "name").Result()
all, _ := rdb.HGetAll(ctx, "user:1").Result()

// HDEL/HEXISTS
rdb.HDel(ctx, "user:1", "name")
exists, _ := rdb.HExists(ctx, "user:1", "name").Result()
```

### List 操作

```go
// LPUSH/RPUSH
rdb.LPush(ctx, "queue", "task1", "task2")
rdb.RPush(ctx, "queue", "task3")

// LPOP/RPOP（阻塞版本 BLPOP/BRPOP）
val, _ := rdb.LPop(ctx, "queue").Result()

// LRANGE/LLEN
vals, _ := rdb.LRange(ctx, "queue", 0, -1).Result()
n, _ := rdb.LLen(ctx, "queue").Result()
```

### Set 操作

```go
// SADD
rdb.SAdd(ctx, "tags", "go", "redis", "database")

// SMEMBERS/SISMEMBER
members, _ := rdb.SMembers(ctx, "tags").Result()
exists, _ := rdb.SIsMember(ctx, "tags", "go").Result()

// 集合运算
inter, _ := rdb.SInter(ctx, "set1", "set2").Result()   // 交集
union, _ := rdb.SUnion(ctx, "set1", "set2").Result()   // 并集
```

### Sorted Set 操作

```go
// ZADD
rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 100, Member: "player1"})

// ZRANGE（按排名）
vals, _ := rdb.ZRange(ctx, "leaderboard", 0, -1).Result()

// ZRANGEBYSCORE（按分数）
vals, _ := rdb.ZRangeByScore(ctx, "leaderboard", &redis.ZRangeBy{
    Min: "-inf", Max: "+inf",
}).Result()

// ZRANK/ZSCORE
rank, _ := rdb.ZRank(ctx, "leaderboard", "player1").Result()
score, _ := rdb.ZScore(ctx, "leaderboard", "player1").Result()
```

### Pipeline（批量操作）

```go
// 基础 Pipeline
pipe := rdb.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Get(ctx, "key1")
cmds, err := pipe.Exec(ctx)

// 处理结果
for i, cmd := range cmds {
    fmt.Printf("cmd %d: %v\n", i, cmd)
}
```

### 事务（WATCH + MULTI/EXEC）

```go
// WATCH 实现乐观锁
err := rdb.Watch(ctx, func(tx *redis.Tx) error {
    // 获取当前值
    n, err := tx.Get(ctx, "counter").Int64()
    if err != nil && err != redis.Nil {
        return err
    }

    // 在事务中执行操作
    _, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        pipe.Set(ctx, "counter", n+1, 0)
        return nil
    })
    return err
}, "counter")
```

### Pub/Sub（发布订阅）

```go
// 订阅频道
pubsub := rdb.Subscribe(ctx, "channel1", "channel2")
defer pubsub.Close()

// 接收消息
ch := pubsub.Channel()
for msg := range ch {
    fmt.Printf("频道：%s, 消息：%s\n", msg.Channel, msg.Payload)
}

// 发布消息
rdb.Publish(ctx, "channel1", "hello")
```

### Streams（流）

```go
// XADD（添加消息）
id, _ := rdb.XAdd(ctx, &redis.XAddArgs{
    Stream: "mystream",
    ID:     "*", // 自动生成ID
    Values: map[string]interface{}{"field1": "value1"},
}).Result()

// 消费者组
rdb.XGroupCreate(ctx, "mystream", "mygroup", "0")
streams, _ := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
    Group:    "mygroup",
    Consumer: "consumer1",
    Streams:  []string{"mystream", ">"},
}).Result()
```

### Lua 脚本

```go
// 定义脚本
script := redis.NewScript(`
    local current = redis.call("GET", KEYS[1])
    if current == false then current = 0 end
    redis.call("SET", KEYS[1], current + ARGV[1])
    return current + ARGV[1]
`)

// 执行脚本
result, err := script.Run(ctx, rdb, []string{"counter"}, 1).Int64()
```

## 实战场景

### 1. 缓存模式（Cache-Aside）

```go
// 获取或设置缓存
func GetOrSet[T any](ctx context.Context, rdb *redis.Client, key string,
    ttl time.Duration, loader func() (T, error)) (T, error) {
    var result T

    // 尝试从缓存获取
    data, err := rdb.Get(ctx, key).Result()
    if err == nil {
        if err := json.Unmarshal([]byte(data), &result); err == nil {
            return result, nil
        }
    }

    // 缓存未命中，从数据源加载
    result, err = loader()
    if err != nil {
        return result, err
    }

    // 写入缓存
    data, _ = json.Marshal(result)
    rdb.Set(ctx, key, data, ttl)

    return result, nil
}
```

### 2. 分布式锁

```go
import "github.com/bsm/redislock"

func withLock(ctx context.Context, rdb *redis.Client, key string,
    ttl time.Duration, fn func() error) error {

    lock, err := redislock.New(rdb).Obtain(ctx, key, ttl, nil)
    if err != nil {
        return fmt.Errorf("获取锁失败: %w", err)
    }
    defer lock.Release(ctx)

    return fn()
}
```

### 3. 限流器

```go
import "github.com/go-redis/redis_rate/v10"

func rateLimitMiddleware(rdb *redis.Client) gin.HandlerFunc {
    limiter := redis_rate.NewLimiter(rdb)

    return func(c *gin.Context) {
        key := "rate:" + c.ClientIP()
        res, _ := limiter.Allow(c, key, redis_rate.PerSecond(10))

        if res.Allowed == 0 {
            c.AbortWithStatus(429)
            return
        }
        c.Next()
    }
}
```

### 4. 会话存储

```go
type SessionStore struct {
    rdb *redis.Client
    ttl time.Duration
}

func (s *SessionStore) Set(ctx context.Context, sessionID string, data map[string]interface{}) error {
    json, _ := json.Marshal(data)
    return s.rdb.Set(ctx, "session:"+sessionID, json, s.ttl).Err()
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (map[string]interface{}, error) {
    data, err := s.rdb.Get(ctx, "session:"+sessionID).Result()
    if err == redis.Nil {
        return nil, nil
    }
    var result map[string]interface{}
    json.Unmarshal([]byte(data), &result)
    return result, err
}
```

### 5. 消息队列

```go
// 生产者
func Enqueue(ctx context.Context, rdb *redis.Client, queue string, task interface{}) error {
    data, _ := json.Marshal(task)
    return rdb.RPush(ctx, queue, data).Err()
}

// 消费者（阻塞弹出）
func Dequeue(ctx context.Context, rdb *redis.Client, queue string, timeout time.Duration) (interface{}, error) {
    result, err := rdb.BRPop(ctx, timeout, queue).Result()
    if err != nil {
        return nil, err
    }
    var task interface{}
    json.Unmarshal([]byte(result[1]), &task)
    return task, nil
}
```

## 高级用法

### Redis Cluster

```go
rdb := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "localhost:7000",
        "localhost:7001",
        "localhost:7002",
    },
    Password: "password",

    // 路由选项
    RouteByLatency: true,  // 按延迟路由
    RouteRandomly:  true,  // 随机路由

    // 连接池
    PoolSize:     100,
    MinIdleConns: 10,
})
```

### Redis Sentinel

```go
rdb := redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{
        "localhost:26379",
        "localhost:26380",
    },
    Password: "password",
    DB:       0,
})
```

### 自定义 Hook（监控/日志）

```go
type MetricsHook struct{}

func (h MetricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
    return func(ctx context.Context, cmd redis.Cmder) error {
        start := time.Now()
        err := next(ctx, cmd)
        duration := time.Since(start)

        // 记录指标
        metrics.RecordRedisCommand(cmd.Name(), duration, err)

        return err
    }
}

rdb.AddHook(MetricsHook{})
```

### OpenTelemetry 集成

```go
import "github.com/redis/go-redis/extra/redisotel/v9"

func main() {
    rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

    // 添加 Tracing 和 Metrics
    errors.Join(
        redisotel.InstrumentTracing(rdb),
        redisotel.InstrumentMetrics(rdb),
    )
}
```

## 框架集成

### Gin 框架

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func init() {
    rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
}

func main() {
    r := gin.Default()

    r.GET("/user/:id", func(c *gin.Context) {
        id := c.Param("id")
        key := "user:" + id

        // 尝试从缓存获取
        data, err := rdb.Get(c, key).Result()
        if err == nil {
            c.Data(200, "application/json", []byte(data))
            return
        }

        // 从数据库获取并缓存
        user, _ := db.GetUser(c, id)
        json, _ := json.Marshal(user)
        rdb.Set(c, key, json, 5*time.Minute)

        c.JSON(200, user)
    })

    r.Run(":8080")
}
```

## 测试

### 使用 miniredis（推荐）

```go
import (
    "testing"
    "github.com/alicebob/miniredis/v2"
    "github.com/redis/go-redis/v9"
)

func TestRedis(t *testing.T) {
    // 创建 mock Redis
    s := miniredis.RunT(t)

    rdb := redis.NewClient(&redis.Options{
        Addr: s.Addr(),
    })

    // 测试
    ctx := context.Background()
    rdb.Set(ctx, "key", "value", 0)

    val, err := rdb.Get(ctx, "key").Result()
    if err != nil || val != "value" {
        t.Errorf("expected value, got %s, err %v", val, err)
    }
}
```

## 错误处理

### 类型化错误检查

```go
func handleError(err error) {
    switch {
    case err == redis.Nil:
        // key 不存在
    case errors.Is(err, context.DeadlineExceeded):
        // 超时
    case redis.IsLoadingError(err):
        // Redis 正在加载
    case redis.IsClusterDownError(err):
        // 集群不可用
    case redis.IsAuthError(err):
        // 认证失败
    case redis.IsOOMError(err):
        // 内存不足
    }
}
```

## 故障排查

### 常见问题

| 问题 | 可能原因 | 解决方案 |
|------|----------|----------|
| 连接超时 | 网络/防火墙/地址错误 | 检查网络、调整 DialTimeout |
| 连接池耗尽 | 并发过高/连接泄漏 | 增大 PoolSize、检查连接关闭 |
| 内存不足 | 无淘汰策略 | 设置 maxmemory-policy |
| 慢查询 | KEYS 命令/大 Key | 使用 SCAN、拆分大 Key |

### 监控连接池

```go
stats := rdb.PoolStats()
fmt.Printf("总连接: %d, 空闲: %d\n",
    stats.TotalConns, stats.IdleConns)
```

## 最佳实践

1. **使用单例模式**管理 Redis 客户端
2. **始终使用 defer 关闭**客户端
3. **传递 Context** 控制超时
4. **使用 Pipeline** 减少 RTT
5. **检查 Nil 错误** 区分不存在和错误
6. **使用冒号分隔**的键命名规范
7. **添加 Hook** 进行监控和日志

## 相关资源

- **官方文档**: https://redis.io/docs/latest/integrate/go-redis/
- **GitHub**: https://github.com/redis/go-redis
- **GoDoc**: https://pkg.go.dev/github.com/redis/go-redis/v9
- **示例**: https://pkg.go.dev/github.com/redis/go-redis/v9#pkg-examples

## 生态系统

- **分布式锁**: https://github.com/bsm/redislock
- **缓存库**: https://github.com/go-redis/cache
- **限流库**: https://github.com/go-redis/redis_rate
- **OpenTelemetry**: `github.com/redis/go-redis/extra/redisotel/v9`
- **测试**: https://github.com/alicebob/miniredis