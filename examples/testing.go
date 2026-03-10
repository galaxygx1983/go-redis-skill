package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// ===========================================
// 使用 miniredis 进行单元测试
// ===========================================

// TestBasicOperations 基础操作测试
func TestBasicOperations(t *testing.T) {
	// 创建 mock Redis 服务器
	s := miniredis.RunT(t)

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	defer rdb.Close()

	ctx := context.Background()

	// 测试 SET/GET
	t.Run("SetGet", func(t *testing.T) {
		err := rdb.Set(ctx, "key", "value", 0).Err()
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// miniredis 提供的辅助方法验证
		s.CheckGet(t, "key", "value")

		// 使用客户端验证
		val, err := rdb.Get(ctx, "key").Result()
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "value" {
			t.Errorf("expected value, got %s", val)
		}
	})

	// 测试 SETNX
	t.Run("SetNX", func(t *testing.T) {
		// 键不存在时设置成功
		ok, err := rdb.SetNX(ctx, "newkey", "newvalue", 0).Result()
		if err != nil || !ok {
			t.Fatalf("SetNX failed: %v, %v", err, ok)
		}

		// 键存在时设置失败
		ok, err = rdb.SetNX(ctx, "newkey", "anothervalue", 0).Result()
		if err != nil || ok {
			t.Fatalf("SetNX should fail for existing key")
		}
	})

	// 测试 TTL
	t.Run("TTL", func(t *testing.T) {
		err := rdb.Set(ctx, "ttlkey", "value", 10*time.Second).Err()
		if err != nil {
			t.Fatalf("Set with TTL failed: %v", err)
		}

		ttl, err := rdb.TTL(ctx, "ttlkey").Result()
		if err != nil {
			t.Fatalf("TTL failed: %v", err)
		}
		if ttl <= 0 || ttl > 10*time.Second {
			t.Errorf("unexpected TTL: %v", ttl)
		}
	})
}

// TestHashOperations Hash 操作测试
func TestHashOperations(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	// 测试 HSET/HGET
	t.Run("HSetHGet", func(t *testing.T) {
		err := rdb.HSet(ctx, "user:1", "name", "张三", "age", 25).Err()
		if err != nil {
			t.Fatalf("HSet failed: %v", err)
		}

		name, err := rdb.HGet(ctx, "user:1", "name").Result()
		if err != nil {
			t.Fatalf("HGet failed: %v", err)
		}
		if name != "张三" {
			t.Errorf("expected 张三, got %s", name)
		}
	})

	// 测试 HGETALL
	t.Run("HGetAll", func(t *testing.T) {
		all, err := rdb.HGetAll(ctx, "user:1").Result()
		if err != nil {
			t.Fatalf("HGetAll failed: %v", err)
		}
		if len(all) != 2 {
			t.Errorf("expected 2 fields, got %d", len(all))
		}
	})
}

// TestListOperations List 操作测试
func TestListOperations(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	// 测试 LPUSH/RPOP
	t.Run("LPushRPop", func(t *testing.T) {
		rdb.LPush(ctx, "list", "a", "b", "c")

		val, err := rdb.RPop(ctx, "list").Result()
		if err != nil {
			t.Fatalf("RPop failed: %v", err)
		}
		if val != "a" {
			t.Errorf("expected a, got %s", val)
		}
	})

	// 测试 LLEN
	t.Run("LLen", func(t *testing.T) {
		len, err := rdb.LLen(ctx, "list").Result()
		if err != nil {
			t.Fatalf("LLen failed: %v", err)
		}
		if len != 2 {
			t.Errorf("expected 2, got %d", len)
		}
	})
}

// TestSortedSetOperations Sorted Set 操作测试
func TestSortedSetOperations(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	t.Run("ZAddZRange", func(t *testing.T) {
		rdb.ZAdd(ctx, "zset", redis.Z{Score: 1, Member: "a"})
		rdb.ZAdd(ctx, "zset", redis.Z{Score: 2, Member: "b"})
		rdb.ZAdd(ctx, "zset", redis.Z{Score: 3, Member: "c"})

		vals, err := rdb.ZRange(ctx, "zset", 0, -1).Result()
		if err != nil {
			t.Fatalf("ZRange failed: %v", err)
		}
		if len(vals) != 3 {
			t.Errorf("expected 3 members, got %d", len(vals))
		}
		if vals[0] != "a" || vals[2] != "c" {
			t.Errorf("unexpected order: %v", vals)
		}
	})
}

// ===========================================
// 接口抽象测试
// ===========================================

// Cache 定义缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// RedisCache Redis 实现
type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// MockCache Mock 实现（用于测试）
type MockCache struct {
	data map[string]string
}

func NewMockCache() *MockCache {
	return &MockCache{data: make(map[string]string)}
}

func (c *MockCache) Get(ctx context.Context, key string) (string, error) {
	val, ok := c.data[key]
	if !ok {
		return "", redis.Nil
	}
	return val, nil
}

func (c *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.data[key] = fmt.Sprintf("%v", value)
	return nil
}

func (c *MockCache) Delete(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}

// TestWithInterface 使用接口测试
func TestWithInterface(t *testing.T) {
	// 使用 Mock 实现
	t.Run("MockCache", func(t *testing.T) {
		cache := NewMockCache()
		testCacheOperations(t, cache)
	})

	// 使用 Redis 实现（需要 miniredis）
	t.Run("RedisCache", func(t *testing.T) {
		s := miniredis.RunT(t)
		rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
		defer rdb.Close()

		cache := NewRedisCache(rdb)
		testCacheOperations(t, cache)
	})
}

func testCacheOperations(t *testing.T, cache Cache) {
	ctx := context.Background()

	// Set
	err := cache.Set(ctx, "testkey", "testvalue", time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get
	val, err := cache.Get(ctx, "testkey")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "testvalue" {
		t.Errorf("expected testvalue, got %s", val)
	}

	// Delete
	err = cache.Delete(ctx, "testkey")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Get after delete
	_, err = cache.Get(ctx, "testkey")
	if err != redis.Nil {
		t.Errorf("expected redis.Nil, got %v", err)
	}
}

// ===========================================
// 集成测试（跳过短测试）
// ===========================================

// TestIntegration 集成测试（需要真实 Redis）
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	ctx := context.Background()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis 不可用")
	}

	// 运行集成测试
	t.Run("RealRedis", func(t *testing.T) {
		// 清理测试数据
		rdb.Del(ctx, "test:integration:key")

		// 测试
		err := rdb.Set(ctx, "test:integration:key", "value", 0).Err()
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		val, err := rdb.Get(ctx, "test:integration:key").Result()
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if val != "value" {
			t.Errorf("expected value, got %s", val)
		}

		// 清理
		rdb.Del(ctx, "test:integration:key")
	})
}

// ===========================================
// 基准测试
// ===========================================

// BenchmarkSetGet SET/GET 性能基准测试
func BenchmarkSetGet(b *testing.B) {
	s, err := miniredis.Run()
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "benchkey", "benchvalue", 0)
		rdb.Get(ctx, "benchkey")
	}
}

// BenchmarkPipeline Pipeline 性能基准测试
func BenchmarkPipeline(b *testing.B) {
	s, err := miniredis.Run()
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipe := rdb.Pipeline()
		for j := 0; j < 10; j++ {
			pipe.Set(ctx, fmt.Sprintf("benchkey%d", j), j, 0)
		}
		pipe.Exec(ctx)
	}
}

// ===========================================
// 辅助函数
// ===========================================

// setupTestRedis 创建测试 Redis 客户端
func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	return rdb, s
}

// teardownTestRedis 清理测试 Redis
func teardownTestRedis(rdb *redis.Client) {
	rdb.Close()
}

// ===========================================
// 示例：测试 JSON 序列化
// ===========================================

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestJSONSerialization(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	// 存储结构体
	user := User{ID: 1, Name: "张三", Age: 25}
	data, _ := json.Marshal(user)

	err := rdb.Set(ctx, "user:1", data, 0).Err()
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 读取并反序列化
	result, err := rdb.Get(ctx, "user:1").Result()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	var retrieved User
	err = json.Unmarshal([]byte(result), &retrieved)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if retrieved.Name != "张三" {
		t.Errorf("expected 张三, got %s", retrieved.Name)
	}
}

// 示例运行方式：
// go test -v ./...
// go test -v -run TestBasicOperations
// go test -short -v ./...  # 跳过集成测试
// go test -bench=. ./...   # 运行基准测试