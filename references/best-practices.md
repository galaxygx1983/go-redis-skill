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
    rdb.Set(ctx, fmt.Sprintf("key%d", i), i, 0)
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

## 5. 键命名规范

### 使用冒号分隔

```go
// ✅ 推荐
user:123:profile
order:456:items
cache:api:users

// ❌ 不推荐
user123
order_456
cache_api_users
```

## 6. 连接池配置

### 生产环境配置

```go
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    
    // 连接池
    PoolSize:     100,
    PoolTimeout:  4 * time.Second,
    MinIdleConns: 10,
    
    // 超时
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    
    // 缓冲区
    ReadBufferSize:  1024 * 1024,
    WriteBufferSize: 1024 * 1024,
})
```

## 7. 监控与日志

### 添加 Hook

```go
type LoggingHook struct{}

func (h LoggingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
    return func(ctx context.Context, cmd redis.Cmder) error {
        start := time.Now()
        err := next(ctx, cmd)
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("Command %s failed: %v (%v)", cmd.Name(), err, duration)
        }
        
        return err
    }
}

rdb.AddHook(LoggingHook{})
```

### 监控连接池

```go
stats := rdb.PoolStats()
fmt.Printf("Total: %d, Idle: %d, Stale: %d\n",
    stats.TotalConns,
    stats.IdleConns,
    stats.StaleConns)
```

## 8. 缓存模式

### Cache-Aside

```go
func getUser(ctx context.Context, id string) (*User, error) {
    key := "user:" + id
    
    // 尝试从缓存获取
    data, err := rdb.Get(ctx, key).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(data), &user)
        return &user, nil
    }
    
    // 缓存未命中，从数据库加载
    user, err := db.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    data, _ := json.Marshal(user)
    rdb.Set(ctx, key, data, 5*time.Minute)
    
    return user, nil
}
```

## 9. 分布式锁

```go
import "github.com/bsm/redislock"

func withLock(ctx context.Context, key string, fn func() error) error {
    client := redislock.New(rdb)
    
    lock, err := client.Obtain(ctx, key, 10*time.Second, nil)
    if err != nil {
        return err
    }
    defer lock.Release(ctx)
    
    return fn()
}
```

## 10. 限流

```go
import "github.com/go-redis/redis_rate/v10"

limiter := redis_rate.NewLimiter(rdb)

res, err := limiter.Allow(ctx, "user:123", redis_rate.PerSecond(10))
if err != nil {
    return err
}

if res.Allowed > 0 {
    // 允许请求
    return handleRequest()
} else {
    // 限流
    return errors.New("rate limit exceeded")
}
```
