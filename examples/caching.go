package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// User 示例数据结构
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// 创建客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("连接失败:", err)
	}
	fmt.Println("✓ Redis 连接成功")

	// 演示各种缓存模式
	basicCacheDemo(rdb)
	cacheAsideDemo(rdb)
	cachePenetrationDemo(rdb)
	cacheAvalancheDemo(rdb)
	sessionStoreDemo(rdb)

	fmt.Println("\n✓ 所有缓存示例执行完成")
}

// 基础缓存示例
func basicCacheDemo(rdb *redis.Client) {
	fmt.Println("\n=== 基础缓存 ===")

	// SET with TTL
	err := rdb.Set(ctx, "cache:user:123", "张三", 5*time.Minute).Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ SET cache:user:123 (TTL: 5分钟)")

	// GET
	val, err := rdb.Get(ctx, "cache:user:123").Result()
	if err == redis.Nil {
		fmt.Println("✗ 缓存不存在")
	} else if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("✓ GET cache:user:123 = %s\n", val)
	}

	// TTL
	ttl, _ := rdb.TTL(ctx, "cache:user:123").Result()
	fmt.Printf("✓ TTL cache:user:123 = %v\n", ttl)

	// DEL
	rdb.Del(ctx, "cache:user:123")
	fmt.Println("✓ DEL cache:user:123")
}

// Cache-Aside 模式示例
func cacheAsideDemo(rdb *redis.Client) {
	fmt.Println("\n=== Cache-Aside 模式 ===")

	// 模拟从数据库加载
	loadFromDB := func(id int) (*User, error) {
		// 模拟数据库查询延迟
		time.Sleep(100 * time.Millisecond)
		return &User{ID: id, Name: fmt.Sprintf("用户%d", id), Age: 25}, nil
	}

	// 获取或设置缓存
	getOrSet := func(ctx context.Context, rdb *redis.Client, key string,
		ttl time.Duration, loader func() (*User, error)) (*User, error) {

		// 尝试从缓存获取
		data, err := rdb.Get(ctx, key).Result()
		if err == nil {
			var user User
			if err := json.Unmarshal([]byte(data), &user); err == nil {
				fmt.Printf("✓ 缓存命中: %s\n", key)
				return &user, nil
			}
		}

		// 缓存未命中，从数据源加载
		fmt.Printf("✗ 缓存未命中: %s，从数据库加载\n", key)
		user, err := loader()
		if err != nil {
			return nil, err
		}

		// 写入缓存
		data, _ = json.Marshal(user)
		rdb.Set(ctx, key, data, ttl)
		fmt.Printf("✓ 写入缓存: %s\n", key)

		return user, nil
	}

	// 第一次调用（缓存未命中）
	user, _ := getOrSet(ctx, rdb, "cache:user:456", 5*time.Minute,
		func() (*User, error) { return loadFromDB(456) })
	fmt.Printf("  结果: %+v\n", user)

	// 第二次调用（缓存命中）
	user, _ = getOrSet(ctx, rdb, "cache:user:456", 5*time.Minute,
		func() (*User, error) { return loadFromDB(456) })
	fmt.Printf("  结果: %+v\n", user)
}

// 防止缓存穿透示例
func cachePenetrationDemo(rdb *redis.Client) {
	fmt.Println("\n=== 防止缓存穿透 ===")

	getWithEmptyCache := func(ctx context.Context, rdb *redis.Client,
		key string, ttl time.Duration) (string, error) {

		val, err := rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			// 缓存不存在，查询数据库
			// 这里模拟数据库查询
			dbVal := "" // 假设数据库也没有

			// 设置空值，防止穿透（使用较短的 TTL）
			rdb.Set(ctx, key, dbVal, ttl)
			fmt.Printf("✓ 设置空值缓存: %s\n", key)
			return dbVal, nil
		}
		return val, err
	}

	// 查询不存在的数据
	val, _ := getWithEmptyCache(ctx, rdb, "cache:nonexistent:1", time.Minute)
	fmt.Printf("  结果: '%s'\n", val)

	// 再次查询（命中空值缓存）
	val, _ = getWithEmptyCache(ctx, rdb, "cache:nonexistent:1", time.Minute)
	fmt.Printf("  结果: '%s'（命中空值缓存）\n", val)
}

// 防止缓存雪崩示例
func cacheAvalancheDemo(rdb *redis.Client) {
	fmt.Println("\n=== 防止缓存雪崩 ===")

	// 设置带随机偏移的 TTL
	setWithRandomTTL := func(ctx context.Context, rdb *redis.Client,
		key string, value interface{}, baseTTL time.Duration) error {

		// 添加随机偏移（0-60秒），避免同时过期
		randomOffset := time.Duration(time.Now().UnixNano()%60) * time.Second
		ttl := baseTTL + randomOffset

		err := rdb.Set(ctx, key, value, ttl).Err()
		fmt.Printf("✓ SET %s (TTL: %v)\n", key, ttl)
		return err
	}

	// 批量设置缓存
	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("cache:avalanche:%d", i)
		setWithRandomTTL(ctx, rdb, key, fmt.Sprintf("值%d", i), 5*time.Minute)
	}
}

// 会话存储示例
func sessionStoreDemo(rdb *redis.Client) {
	fmt.Println("\n=== 会话存储 ===")

	// 创建会话
	sessionID := "session:abc123"
	sessionData := map[string]interface{}{
		"user_id":  123,
		"username": "张三",
		"role":     "admin",
		"login_at": time.Now().Unix(),
	}

	// 存储会话
	data, _ := json.Marshal(sessionData)
	rdb.Set(ctx, sessionID, data, 30*time.Minute)
	fmt.Printf("✓ 创建会话: %s\n", sessionID)

	// 获取会话
	data, _ = rdb.Get(ctx, sessionID).Result()
	var retrieved map[string]interface{}
	json.Unmarshal(data, &retrieved)
	fmt.Printf("✓ 获取会话: %+v\n", retrieved)

	// 延长会话 TTL
	rdb.Expire(ctx, sessionID, 30*time.Minute)
	fmt.Println("✓ 延长会话 TTL")

	// 删除会话（登出）
	rdb.Del(ctx, sessionID)
	fmt.Println("✓ 删除会话（登出）")
}