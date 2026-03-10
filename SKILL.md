---
name: go-redis
description: Redis 官方 Go 客户端库开发指南。22k stars，BSD-2-Clause 许可。适用于：(1) Go 项目集成 Redis (2) 连接池/集群/Sentinel 配置 (3) Pipeline/事务/PubSub 开发 (4) 性能优化与监控 (5) OpenTelemetry 集成
---

# go-redis 开发指南

> **官方 Redis Go 客户端** - 22k stars，类型安全，功能完整

## 快速开始

### 1. 安装

```bash
# 初始化 Go 模块
go mod init github.com/my/repo

# 安装 go-redis v9
go get github.com/redis/go-redis/v9
```

### 2. 基础连接

```go
package main

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
    // 创建客户端
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // 无密码
        DB:       0,  // 默认数据库
    })
    defer rdb.Close()

    // 测试连接
    err := rdb.Ping(ctx).Err()
    if err != nil {
        panic(err)
    }

    // 设置键值
    err = rdb.Set(ctx, "key", "value", 0).Err()
    if err != nil {
        panic(err)
    }

    // 获取键值
    val, err := rdb.Get(ctx, "key").Result()
    if err != nil {
        panic(err)
    }
    fmt.Println("key:", val)

    // 处理 Nil 错误
    val2, err := rdb.Get(ctx, "key2").Result()
    if err == redis.Nil {
        fmt.Println("key2 不存在")
    } else if err != nil {
        panic(err)
    } else {
        fmt.Println("key2:", val2)
    }
}
```

### 3. 使用 Redis URL 连接

```go
import "github.com/redis/go-redis/v9"

func createClient() *redis.Client {
    url := "redis://user:password@localhost:6379/0?protocol=3"
    opts, err := redis.ParseURL(url)
    if err != nil {
        panic(err)
    }
    return redis.NewClient(opts)
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
    
    // 数据库
    DB: 0,                      // 数据库编号
    
    // 协议版本
    Protocol: 3,                // 3=RESP3, 2=RESP2
    
    // TLS 配置
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
})
```

### 连接池配置

```go
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    
    // 连接池大小
    PoolSize:     100,          // 最大连接数
    PoolTimeout:  4 * time.Second, // 连接池超时
    MinIdleConns: 10,           // 最小空闲连接
    
    // 连接超时
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    
    // 缓冲区大小（v9.12+ 默认 32KiB）
    ReadBufferSize:  1024 * 1024, // 1MB
    WriteBufferSize: 1024 * 1024, // 1MB
})
```

### 认证方式（优先级从高到低）

```go
// 1. 流式凭证提供者（实验性，支持动态更新）
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    StreamingCredentialsProvider: &MyCredentialsProvider{},
})

// 2. 基于上下文的凭证提供者
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    CredentialsProviderContext: func(ctx context.Context) (string, string, error) {
        return "user", "pass", nil
    },
})

// 3. 普通凭证提供者
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    CredentialsProvider: func() (string, string) {
        return "user", "pass"
    },
})

// 4. 用户名/密码字段（最低优先级）
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Username: "user",
    Password: "pass",
})
```

## 核心功能

### 字符串操作

```go
// SET
err := rdb.Set(ctx, "key", "value", 0).Err()

// SET with expiration
err := rdb.Set(ctx, "key", "value", 10*time.Second).Err()

// SETNX
set, err := rdb.SetNX(ctx, "key", "value", 10*time.Second).Result()

// GET
val, err := rdb.Get(ctx, "key").Result()

// GETEX
val, err := rdb.GetEx(ctx, "key", 10*time.Second).Result()

// MSET/MGET
err := rdb.MSet(ctx, map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
}).Err()

vals, err := rdb.MGet(ctx, "key1", "key2").Result()
```

### Hash 操作

```go
// HSET
err := rdb.HSet(ctx, "hash", "field1", "value1").Err()

// HGET
val, err := rdb.HGet(ctx, "hash", "field1").Result()

// HGETALL
all, err := rdb.HGetAll(ctx, "hash").Result()

// HMSET
err := rdb.HMSet(ctx, "hash", map[string]interface{}{
    "field1": "value1",
    "field2": "value2",
}).Err()

// HDEL
err := rdb.HDel(ctx, "hash", "field1").Err()
```

### List 操作

```go
// LPUSH/RPUSH
err := rdb.LPush(ctx, "list", "value1", "value2").Err()
err := rdb.RPush(ctx, "list", "value3").Err()

// LPOP/RPOP
val, err := rdb.LPop(ctx, "list").Result()
val, err := rdb.RPop(ctx, "list").Result()

// LRANGE
vals, err := rdb.LRange(ctx, "list", 0, -1).Result()

// LLEN
length, err := rdb.LLen(ctx, "list").Result()
```

### Set 操作

```go
// SADD
err := rdb.SAdd(ctx, "set", "member1", "member2").Err()

// SMEMBERS
members, err := rdb.SMembers(ctx, "set").Result()

// SISMEMBER
exists, err := rdb.SIsMember(ctx, "set", "member1").Result()

// SREM
err := rdb.SRem(ctx, "set", "member1").Err()

// SINTER/SUNION/SDIFF
inter, err := rdb.SInter(ctx, "set1", "set2").Result()
union, err := rdb.SUnion(ctx, "set1", "set2").Result()
diff, err := rdb.SDiff(ctx, "set1", "set2").Result()
```

### Sorted Set 操作

```go
// ZADD
err := rdb.ZAdd(ctx, "zset", redis.Z{
    Score:  1.0,
    Member: "member1",
}).Err()

// ZRANGE
vals, err := rdb.ZRange(ctx, "zset", 0, -1).Result()

// ZRANGEBYSCORE
vals, err := rdb.ZRangeByScore(ctx, "zset", &redis.ZRangeBy{
    Min: "-inf",
    Max: "+inf",
}).Result()

// ZREM
err := rdb.ZRem(ctx, "zset", "member1").Err()

// ZCARD
count, err := rdb.ZCard(ctx, "zset").Result()
```

### Pipeline（管道）

```go
// 基础 Pipeline
pipe := rdb.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Get(ctx, "key1")
pipe.Get(ctx, "key2")

cmds, err := pipe.Exec(ctx)
if err != nil && err != redis.Nil {
    panic(err)
}

// 处理每个命令的结果
for _, cmd := range cmds {
    fmt.Println(cmd.String())
}

// TxPipeline（事务管道）
pipe := rdb.TxPipeline()
pipe.Multi()
pipe.Incr(ctx, "counter")
pipe.Incr(ctx, "counter")
pipe.Exec()
_, err := pipe.Exec(ctx)
```

### 事务（Transaction）

```go
// WATCH + MULTI + EXEC
err := rdb.Watch(ctx, func(tx *redis.Tx) error {
    // 读取当前值
    n, err := tx.Get(ctx, "counter").Int64()
    if err != nil && err != redis.Nil {
        return err
    }

    // 执行操作
    _, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        pipe.Set(ctx, "counter", n+1, 0)
        return nil
    })
    return err
}, "counter")

if err != nil {
    panic(err)
}
```

### Pub/Sub（发布订阅）

```go
// 订阅频道
pubsub := rdb.Subscribe(ctx, "channel1", "channel2")

// 接收消息
ch := pubsub.Channel()
for msg := range ch {
    fmt.Printf("频道：%s, 消息：%s\n", msg.Channel, msg.Payload)
}

// 发布消息
err := rdb.Publish(ctx, "channel1", "hello").Err()

// 取消订阅
pubsub.Close()

// 带重连的持久化订阅
pubsub := rdb.PSubscribe(ctx, "pattern*")
```

### Streams（流）

```go
// XADD
id, err := rdb.XAdd(ctx, &redis.XAddArgs{
    Stream: "mystream",
    ID:     "*",
    Values: map[string]interface{}{
        "field1": "value1",
        "field2": "value2",
    },
}).Result()

// XREAD
streams, err := rdb.XRead(ctx, &redis.XReadArgs{
    Streams: []string{"mystream", "0"}, // 从 ID=0 开始
    Count:   10,
    Block:   time.Second,
}).Result()

// XGROUP CREATE
err := rdb.XGroupCreate(ctx, "mystream", "mygroup", "0").Err()

// XREADGROUP
streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
    Group:    "mygroup",
    Consumer: "consumer1",
    Streams:  []string{"mystream", ">"},
    Count:    10,
    Block:    time.Second,
}).Result()

// XACK
err := rdb.XAck(ctx, "mystream", "mygroup", "message-id").Err()
```

### Lua 脚本

```go
// 定义脚本
script := redis.NewScript(`
    redis.call("SET", KEYS[1], ARGV[1])
    return redis.call("GET", KEYS[1])
`)

// 执行脚本
result, err := script.Run(ctx, rdb, []string{"key"}, "value").Result()

// 使用 EVALSHA（自动回退到 EVAL）
result, err := script.Run(ctx, rdb, []string{"key"}, "value").Result()

// 并发安全
var mu sync.Mutex
script.Run(ctx, rdb, []string{"key"}, "value")
```

## 高级用法

### Redis Cluster

```go
import "github.com/redis/go-redis/v9"

func createClusterClient() *redis.ClusterClient {
    rdb := redis.NewClusterClient(&redis.ClusterOptions{
        Addrs: []string{
            "localhost:7000",
            "localhost:7001",
            "localhost:7002",
        },
        Password: "password", // 可选
        
        // 路由选项
        RouteByLatency: true,  // 按延迟路由
        RouteRandomly:  true,  // 随机路由
        
        // 连接池
        PoolSize:     100,
        MinIdleConns: 10,
    })
    return rdb
}

// 使用 Cluster 客户端
rdb := createClusterClient()
defer rdb.Close()

// 对特定节点执行命令
err := rdb.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
    return shard.Ping(ctx).Err()
}).Err()
```

### Redis Sentinel

```go
import "github.com/redis/go-redis/v9"

func createSentinelClient() *redis.Client {
    rdb := redis.NewFailoverClient(&redis.FailoverOptions{
        MasterName:    "mymaster",
        SentinelAddrs: []string{
            "localhost:26379",
            "localhost:26380",
            "localhost:26381",
        },
        Password: "password",
        DB:       0,
        
        // 连接池
        PoolSize:     100,
        MinIdleConns: 10,
    })
    return rdb
}
```

### 自定义 Hook

```go
// 日志 Hook
type LoggingHook struct{}

func (h LoggingHook) DialHook(next redis.DialHook) redis.DialHook {
    return func(ctx context.Context, network, addr string) (net.Conn, error) {
        log.Printf("Dialing %s %s", network, addr)
        return next(ctx, network, addr)
    }
}

func (h LoggingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
    return func(ctx context.Context, cmd redis.Cmder) error {
        start := time.Now()
        err := next(ctx, cmd)
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("Command %s failed: %v (took %v)", cmd.Name(), err, duration)
        } else {
            log.Printf("Command %s succeeded (took %v)", cmd.Name(), duration)
        }
        
        return err
    }
}

func (h LoggingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
    return func(ctx context.Context, cmds []redis.Cmder) error {
        start := time.Now()
        err := next(ctx, cmds)
        duration := time.Since(start)
        
        log.Printf("Pipeline executed %d commands in %v", len(cmds), duration)
        return err
    }
}

// 注册 Hook
rdb.AddHook(LoggingHook{})
```

### 错误处理

```go
import "github.com/redis/go-redis/v9"

// 类型化错误检查
func handleError(err error) {
    if err == nil {
        return
    }
    
    // 连接错误
    if errors.Is(err, context.DeadlineExceeded) {
        // 超时处理
    }
    
    // Redis 特定错误
    if redis.IsLoadingError(err) {
        // Redis 正在加载数据
    }
    if redis.IsReadOnlyError(err) {
        // 写入只读副本
    }
    if redis.IsClusterDownError(err) {
        // 集群不可用
    }
    if redis.IsAuthError(err) {
        // 认证失败
    }
    if redis.IsPermissionError(err) {
        // 权限不足
    }
    if redis.IsOOMError(err) {
        // 内存不足
    }
    
    // 自定义错误包装
    type AppError struct {
        Code string
        Err  error
    }
    
    func (e *AppError) Error() string {
        return fmt.Sprintf("[%s] %v", e.Code, e.Err)
    }
    
    func (e *AppError) Unwrap() error {
        return e.Err
    }
    
    // 使用
    wrappedErr := &AppError{Code: "REDIS_ERROR", Err: err}
    cmd.SetErr(wrappedErr)
}
```

### OpenTelemetry 集成

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/redis/go-redis/extra/redisotel/v9"
)

func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 添加 Tracing 和 Metrics
    if err := errors.Join(
        redisotel.InstrumentTracing(rdb),
        redisotel.InstrumentMetrics(rdb),
    ); err != nil {
        log.Fatal(err)
    }
}
```

## 最佳实践

### 1. 连接管理

```go
// ✅ 正确：使用 defer 关闭
rdb := redis.NewClient(&redis.Options{...})
defer rdb.Close()

// ✅ 正确：使用单例模式
var redisClient *redis.Client
var once sync.Once

func GetRedisClient() *redis.Client {
    once.Do(func() {
        redisClient = redis.NewClient(&redis.Options{...})
    })
    return redisClient
}

// ❌ 错误：每次调用都创建新客户端
func bad() {
    rdb := redis.NewClient(...) // 资源泄漏
}
```

### 2. Context 使用

```go
// ✅ 正确：传递 Context
func getUser(ctx context.Context, id string) (string, error) {
    return rdb.Get(ctx, "user:"+id).Result()
}

// ✅ 正确：设置超时
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
val, err := rdb.Get(ctx, "key").Result()

// ❌ 错误：使用 context.Background() 在函数内部
func bad() {
    ctx := context.Background() // 无法控制超时
    rdb.Get(ctx, "key")
}
```

### 3. Pipeline 批量操作

```go
// ✅ 正确：使用 Pipeline 减少 RTT
pipe := rdb.Pipeline()
for i := 0; i < 100; i++ {
    pipe.Set(ctx, fmt.Sprintf("key%d", i), i, 0)
}
_, err := pipe.Exec(ctx)

// ❌ 错误：逐个执行
for i := 0; i < 100; i++ {
    rdb.Set(ctx, fmt.Sprintf("key%d", i), i, 0) // 100 次 RTT
}
```

### 4. 错误处理

```go
// ✅ 正确：检查 Nil 错误
val, err := rdb.Get(ctx, "key").Result()
if err == redis.Nil {
    // key 不存在
    return "", nil
} else if err != nil {
    return "", err
}

// ❌ 错误：忽略错误
val, _ := rdb.Get(ctx, "key").Result() // 可能获取到错误值
```

### 5. 键命名规范

```go
// ✅ 推荐：使用冒号分隔的命名
user:123:profile
order:456:items
cache:api:users

// ❌ 不推荐：无规律的命名
user123
order_456
cache_api_users
```

## 常见问题

### Q: 如何处理连接超时？

```go
rdb := redis.NewClient(&redis.Options{
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
})

// 使用带超时的 Context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := rdb.Ping(ctx).Err()
```

### Q: 如何实现分布式锁？

```go
import "github.com/bsm/redislock"

func distributedLock() {
    client := redislock.New(rdb)
    
    lock, err := client.Obtain(ctx, "my-lock", 10*time.Second, nil)
    if err != nil {
        // 获取锁失败
        return
    }
    defer lock.Release(ctx)
    
    // 执行临界区代码
}
```

### Q: 如何实现缓存？

```go
import "github.com/go-redis/cache/v9"

func caching() {
    c := cache.New(&cache.Options{
        Redis:      rdb,
        LocalCache: cache.NewTinyLFU(1000, time.Minute),
    })
    
    var data MyStruct
    err := c.Get(ctx, "my-key", &data)
    if err == cache.ErrCacheMiss {
        // 缓存未命中，从数据库加载
        data = loadData()
        c.Set(ctx, "my-key", data, 5*time.Minute)
    }
}
```

### Q: 如何实现限流？

```go
import "github.com/go-redis/redis_rate/v10"

func rateLimiting() {
    limiter := redis_rate.NewLimiter(rdb)
    
    res, err := limiter.Allow(ctx, "user:123", redis_rate.PerSecond(10))
    if err != nil {
        panic(err)
    }
    
    if res.Allowed > 0 {
        // 允许请求
    } else {
        // 限流
    }
}
```

## 相关资源

- **官方文档**: https://redis.io/docs/latest/integrate/go-redis/
- **GitHub**: https://github.com/redis/go-redis
- **GoDoc**: https://pkg.go.dev/github.com/redis/go-redis/v9
- **示例**: https://pkg.go.dev/github.com/redis/go-redis/v9#pkg-examples
- **Discord**: https://discord.gg/W4txy5AeKM

## 生态系统

- **分布式锁**: https://github.com/bsm/redislock
- **缓存库**: https://github.com/go-redis/cache
- **限流库**: https://github.com/go-redis/redis_rate
- **Entra ID 认证**: https://github.com/redis/go-redis-entraid
