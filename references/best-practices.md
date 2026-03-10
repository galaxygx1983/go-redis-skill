# go-redis 最佳实践

## 1. 连接管理

### 使用单例模式

```go
var (
    redisClient *redis.Client
    once        sync.Once
)

func GetRedisClient() *redis.Client {
    once.Do(func() {
        redisClient = redis.NewClient(&redis.Options{
            Addr:     "localhost:6379",
            PoolSize: 100,
        })
    })
    return redisClient
}
```

### 始终关闭客户端

```go
func main() {
    rdb := redis.NewClient(&redis.Options{...})
    defer rdb.Close() // 确保关闭
}
```

### 生产环境配置

```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    PoolSize: runtime.GOMAXPROCS(0) * 10, // CPU核心数 * 10

    // 超时设置
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,

    // 连接池
    PoolSize:        100,
    MinIdleConns:    10,
    PoolTimeout:     4 * time.Second,
    ConnMaxLifetime: 30 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,

    // 缓冲区（大数据量场景）
    ReadBufferSize:  64 * 1024, // 64KB
    WriteBufferSize: 64 * 1024,
})
```

## 2. Context 使用

### 传递 Context

```go
// ✅ 推荐
func getUser(ctx context.Context, id string) (string, error) {
    return rdb.Get(ctx, "user:"+id).Result()
}

// ❌ 不推荐
func getUser(id string) (string, error) {
    ctx := context.Background() // 无法控制超时
    return rdb.Get(ctx, "user:"+id).Result()
}
```

### 设置超时

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

val, err := rdb.Get(ctx, "key").Result()
if err == context.DeadlineExceeded {
    // 处理超时
}
```

## 3. 错误处理

### 检查 Nil 错误

```go
val, err := rdb.Get(ctx, "key").Result()
if err == redis.Nil {
    // key 不存在
    return "", nil
} else if err != nil {
    return "", err
}
```

### 类型化错误检查

```go
if redis.IsLoadingError(err) {
    // Redis 正在加载
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
if redis.IsClusterDownError(err) {
    // 集群不可用
}
```

### 错误包装

```go
type RedisError struct {
    Op  string
    Key string
    Err error
}

func (e *RedisError) Error() string {
    return fmt.Sprintf("redis %s %s: %v", e.Op, e.Key, e.Err)
}

func (e *RedisError) Unwrap() error {
    return e.Err
}

// 使用
val, err := rdb.Get(ctx, "key").Result()
if err != nil {
    return nil, &RedisError{Op: "GET", Key: "key", Err: err}
}
```

## 4. 性能优化

### 使用 Pipeline

```go
// ✅ 推荐：使用 Pipeline
pipe := rdb.Pipeline()
for i := 0; i < 100; i++ {
    pipe.Set(ctx, fmt.Sprintf("key%d", i), i, 0)
}
_, err := pipe.Exec(ctx)

// ❌ 不推荐：逐个执行
for i := 0; i < 100; i++ {
    rdb.Set(ctx, fmt.Sprintf("key%d", i), i, 0) // 100 次 RTT
}
```

### 批量操作

```go
// MSET 代替多个 SET
rdb.MSet(ctx, map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
    "key3": "value3",
})

// MGET 代替多个 GET
vals, err := rdb.MGet(ctx, "key1", "key2", "key3").Result()
```

### 使用 Lua 脚本

```go
// 合并多个操作为一个原子操作
script := redis.NewScript(`
    local current = redis.call("GET", KEYS[1])
    if current == false then
        current = 0
    end
    redis.call("SET", KEYS[1], current + ARGV[1])
    return current + ARGV[1]
`)
result, err := script.Run(ctx, rdb, []string{"counter"}, 1).Int64()
```

### 使用 SCAN 代替 KEYS

```go
// ✅ 推荐：使用 SCAN
iter := rdb.Scan(ctx, 0, "prefix:*", 0).Iterator()
for iter.Next(ctx) {
    key := iter.Val()
    // 处理 key
}

// ❌ 避免：使用 KEYS（会阻塞）
keys, _ := rdb.Keys(ctx, "prefix:*").Result()
```

## 5. 键命名规范

### 使用冒号分隔

```go
// ✅ 推荐：冒号分隔，便于管理和分组
user:123:profile
order:456:items
cache:api:users
session:abc123

// ❌ 不推荐：无规律的命名
user123
order_456
cache_api_users
```

### 使用前缀隔离

```go
const (
    keyPrefix    = "myapp:"
    userPrefix   = keyPrefix + "user:"
    cachePrefix  = keyPrefix + "cache:"
    sessionPrefix = keyPrefix + "session:"
)

func userKey(id string) string {
    return userPrefix + id
}
```

## 6. 监控与日志

### 添加 Hook

```go
type MetricsHook struct{}

func (h MetricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
    return func(ctx context.Context, cmd redis.Cmder) error {
        start := time.Now()
        err := next(ctx, cmd)
        duration := time.Since(start)

        // 记录指标
        metrics.RedisCommands.Inc()
        metrics.RedisLatency.Observe(duration.Seconds())

        if err != nil && err != redis.Nil {
            metrics.RedisErrors.Inc()
            log.Printf("Redis command %s failed: %v (%v)", cmd.Name(), err, duration)
        }

        return err
    }
}

rdb.AddHook(MetricsHook{})
```

### 监控连接池

```go
func monitorPool(rdb *redis.Client) {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        stats := rdb.PoolStats()
        log.Printf("连接池: 总=%d, 空闲=%d, 等待=%d",
            stats.TotalConns,
            stats.IdleConns,
            stats.WaitCount)
    }
}
```

## 7. 缓存模式

### Cache-Aside 模式

```go
func GetOrSet[T any](ctx context.Context, rdb *redis.Client,
    key string, ttl time.Duration, loader func() (T, error)) (T, error) {
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

    // 写入缓存（异步，不阻塞）
    go func() {
        data, _ := json.Marshal(result)
        rdb.Set(context.Background(), key, data, ttl)
    }()

    return result, nil
}
```

### 防止缓存穿透

```go
func GetWithEmptyCache(ctx context.Context, rdb *redis.Client,
    key string, ttl time.Duration) (string, error) {
    val, err := rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        // 设置空值，防止穿透
        rdb.Set(ctx, key, "", ttl)
        return "", nil
    }
    return val, err
}
```

### 防止缓存雪崩

```go
func SetWithRandomTTL(ctx context.Context, rdb *redis.Client,
    key string, value interface{}, baseTTL time.Duration) error {
    // 添加随机偏移，避免同时过期
    randomOffset := time.Duration(rand.Intn(60)) * time.Second
    return rdb.Set(ctx, key, value, baseTTL+randomOffset).Err()
}
```

## 8. 分布式锁

```go
import "github.com/bsm/redislock"

func withLock(ctx context.Context, rdb *redis.Client,
    key string, ttl time.Duration, fn func() error) error {

    client := redislock.New(rdb)

    lock, err := client.Obtain(ctx, key, ttl, nil)
    if err != nil {
        return fmt.Errorf("获取锁失败: %w", err)
    }
    defer lock.Release(ctx)

    return fn()
}

// 使用
err := withLock(ctx, rdb, "order:123", 10*time.Second, func() error {
    // 执行临界区代码
    return nil
})
```

## 9. 限流

```go
import "github.com/go-redis/redis_rate/v10"

func rateLimitMiddleware(rdb *redis.Client) gin.HandlerFunc {
    limiter := redis_rate.NewLimiter(rdb)

    return func(c *gin.Context) {
        key := "rate:" + c.ClientIP()

        res, err := limiter.Allow(c, key, redis_rate.PerSecond(10))
        if err != nil {
            c.AbortWithStatus(500)
            return
        }

        if res.Allowed == 0 {
            c.Header("X-RateLimit-Remaining", "0")
            c.AbortWithStatus(429)
            return
        }

        c.Header("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
        c.Next()
    }
}
```

## 10. 安全配置

### TLS 连接

```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Username: "user",
    Password: "password",
    TLSConfig: &tls.Config{
        MinVersion:         tls.VersionTLS12,
        InsecureSkipVerify: false, // 生产环境必须验证证书
    },
})
```

### ACL 用户管理（Redis 6.0+）

```go
// 使用 ACL 用户
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Username: "app_user", // ACL 用户
    Password: "secure_password",
})
```

### 敏感信息保护

```go
// 使用环境变量
rdb := redis.NewClient(&redis.Options{
    Addr:     os.Getenv("REDIS_ADDR"),
    Username: os.Getenv("REDIS_USER"),
    Password: os.Getenv("REDIS_PASSWORD"),
})

// 或使用凭证提供者（支持动态更新）
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    CredentialsProviderContext: func(ctx context.Context) (string, string, error) {
        // 从密钥管理服务获取凭证
        return secrets.GetRedisCredentials(ctx)
    },
})
```

## 11. 测试

### 使用 miniredis

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

    ctx := context.Background()

    // 测试 SET/GET
    rdb.Set(ctx, "key", "value", 0)
    s.CheckGet(t, "key", "value")

    val, err := rdb.Get(ctx, "key").Result()
    if err != nil {
        t.Fatal(err)
    }
    if val != "value" {
        t.Errorf("expected value, got %s", val)
    }
}
```

### 接口抽象

```go
// 定义接口
type Cache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

// Redis 实现
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}

// Mock 实现
type MockCache struct {
    data map[string]string
}

func (c *MockCache) Get(ctx context.Context, key string) (string, error) {
    val, ok := c.data[key]
    if !ok {
        return "", redis.Nil
    }
    return val, nil
}
```

## 12. 故障排查清单

| 问题 | 可能原因 | 解决方案 |
|------|----------|----------|
| 连接超时 | 网络/防火墙/地址错误 | 检查网络、调整 DialTimeout |
| 连接池耗尽 | 并发过高/连接泄漏 | 增大 PoolSize、检查连接关闭 |
| 内存不足 | 无淘汰策略 | 设置 maxmemory-policy |
| 慢查询 | KEYS 命令/大 Key | 使用 SCAN、拆分大 Key |
| 集群不可用 | 节点故障/网络分区 | 检查集群状态、重试机制 |
| 认证失败 | 密码错误/ACL 配置 | 检查用户名密码、ACL 规则 |