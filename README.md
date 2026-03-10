# go-redis Skill

> Redis 官方 Go 客户端 - 22k stars，类型安全，功能完整

## 技能描述

本 skill 提供 **redis/go-redis**（官方 Redis Go 客户端）的完整开发指南，包括：

- ✅ 基础连接与配置
- ✅ 所有 Redis 数据类型操作
- ✅ Pipeline/事务/PubSub
- ✅ Cluster/Sentinel 支持
- ✅ 连接池管理
- ✅ 性能优化与监控
- ✅ OpenTelemetry 集成
- ✅ 错误处理
- ✅ 实战场景模式（缓存、分布式锁、限流、会话、消息队列）
- ✅ 框架集成（Gin、Wire）
- ✅ 测试指南（miniredis）

## 触发条件

当用户需要：
1. Go 项目集成 Redis
2. 配置 Redis 连接池/集群/Sentinel
3. 实现 Redis 缓存/分布式锁/限流
4. Redis 性能调优和问题排查
5. 使用 go-redis 库进行开发

## 文件结构

```
go-redis-skill/
├── SKILL.md              # 主技能文档（完整指南）
├── README.md             # 本文件
├── examples/
│   ├── basic_usage.go    # 基础操作示例
│   ├── connection_pool.go # 连接池管理
│   ├── cluster_usage.go  # Cluster 使用
│   ├── caching.go        # 缓存模式
│   ├── distributed_lock.go # 分布式锁
│   └── testing.go        # 测试示例
└── references/
    └── best-practices.md # 最佳实践
```

## 快速开始

### 安装

```bash
go get github.com/redis/go-redis/v9
```

### 基础使用

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})
defer rdb.Close()

// SET/GET
rdb.Set(ctx, "key", "value", 0)
val, _ := rdb.Get(ctx, "key").Result()
```

## 核心功能速查

| 功能 | 方法 | 示例 |
|------|------|------|
| 字符串 | Set, Get, MSet, MGet | `rdb.Set(ctx, "key", "value", 0)` |
| Hash | HSet, HGet, HGetAll | `rdb.HSet(ctx, "hash", "field", "value")` |
| List | LPush, RPop, LRange | `rdb.LPush(ctx, "list", "item")` |
| Set | SAdd, SMembers, SInter | `rdb.SAdd(ctx, "set", "member")` |
| Sorted Set | ZAdd, ZRange, ZRem | `rdb.ZAdd(ctx, "zset", redis.Z{Score: 1, Member: "m"})` |
| Pipeline | Pipeline() | `pipe.Exec(ctx)` |
| 事务 | Watch() | `rdb.Watch(ctx, keys...)` |
| PubSub | Subscribe() | `rdb.Subscribe(ctx, channels...)` |
| Streams | XAdd, XRead | `rdb.XAdd(ctx, &redis.XAddArgs{...})` |
| Lua脚本 | NewScript() | `script.Run(ctx, rdb, keys, args)` |

## 高级功能

- **Pipeline**: `rdb.Pipeline()`
- **事务**: `rdb.Watch(ctx, keys...)`
- **PubSub**: `rdb.Subscribe(ctx, channels...)`
- **Streams**: `rdb.XAdd()`, `rdb.XRead()`
- **Lua 脚本**: `redis.NewScript()`
- **Cluster**: `redis.NewClusterClient()`
- **Sentinel**: `redis.NewFailoverClient()`

## 生产配置

```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    PoolSize: runtime.GOMAXPROCS(0) * 10, // CPU核心数 * 10

    // 超时
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,

    // 连接池
    MinIdleConns:    10,
    ConnMaxLifetime: 30 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
})
```

## 最佳实践

1. **使用单例模式**管理 Redis 客户端
2. **始终使用 defer 关闭**客户端
3. **传递 Context** 控制超时
4. **使用 Pipeline** 减少 RTT
5. **检查 Nil 错误** 区分不存在和错误
6. **使用冒号分隔**的键命名规范
7. **添加 Hook** 进行监控和日志

详细最佳实践见 `references/best-practices.md`

## 错误处理

```go
val, err := rdb.Get(ctx, "key").Result()
if err == redis.Nil {
    // key 不存在
} else if err != nil {
    // 其他错误
}
```

类型化错误检查：

```go
redis.IsLoadingError(err)    // Redis 正在加载
redis.IsAuthError(err)       // 认证失败
redis.IsPermissionError(err) // 权限不足
redis.IsOOMError(err)        // 内存不足
```

## 故障排查

详见 SKILL.md 中的"故障排查"章节。

## 生态系统

- **分布式锁**: https://github.com/bsm/redislock
- **缓存库**: https://github.com/go-redis/cache
- **限流库**: https://github.com/go-redis/redis_rate
- **OpenTelemetry**: `github.com/redis/go-redis/extra/redisotel/v9`
- **测试**: https://github.com/alicebob/miniredis

## 相关资源

- **官方文档**: https://redis.io/docs/latest/integrate/go-redis/
- **GitHub**: https://github.com/redis/go-redis
- **GoDoc**: https://pkg.go.dev/github.com/redis/go-redis/v9
- **示例**: https://pkg.go.dev/github.com/redis/go-redis/v9#pkg-examples

## 许可证

BSD-2-Clause License