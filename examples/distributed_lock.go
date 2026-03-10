package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

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

	// 演示各种分布式锁实现
	simpleLockDemo(rdb)
	luaLockDemo(rdb)
	rateLimiterDemo(rdb)
	messageQueueDemo(rdb)

	fmt.Println("\n✓ 所有分布式锁示例执行完成")
}

// 简单锁示例（SET NX 实现）
func simpleLockDemo(rdb *redis.Client) {
	fmt.Println("\n=== 简单锁（SET NX） ===")

	lockKey := "lock:simple:resource"
	lockValue := "locked"

	// 尝试获取锁
	acquired, err := rdb.SetNX(ctx, lockKey, lockValue, 10*time.Second).Result()
	if err != nil {
		log.Fatal(err)
	}

	if acquired {
		fmt.Println("✓ 获取锁成功")
		// 执行临界区代码
		time.Sleep(100 * time.Millisecond)
		// 释放锁
		rdb.Del(ctx, lockKey)
		fmt.Println("✓ 释放锁")
	} else {
		fmt.Println("✗ 获取锁失败（锁已被占用）")
	}
}

// Lua 脚本锁示例（原子释放）
func luaLockDemo(rdb *redis.Client) {
	fmt.Println("\n=== Lua 脚本锁（原子释放） ===")

	// Lua 脚本：只有值匹配时才删除
	unlockScript := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		end
		return 0
	`)

	// Lua 脚本：带自动续期的锁
	renewScript := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		end
		return 0
	`)

	lockKey := "lock:lua:resource"
	lockValue := fmt.Sprintf("lock_%d", time.Now().UnixNano())

	// 获取锁
	acquired, err := rdb.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()
	if err != nil {
		log.Fatal(err)
	}

	if acquired {
		fmt.Println("✓ 获取锁成功")

		// 模拟执行临界区代码
		for i := 0; i < 3; i++ {
			time.Sleep(2 * time.Second)
			// 续期
			result, _ := renewScript.Run(ctx, rdb, []string{lockKey}, lockValue, 5000).Int64()
			if result == 1 {
				fmt.Printf("✓ 锁续期成功 (第 %d 次)\n", i+1)
			} else {
				fmt.Printf("✗ 锁续期失败 (锁可能已被释放)\n")
				break
			}
		}

		// 释放锁（原子操作）
		result, _ := unlockScript.Run(ctx, rdb, []string{lockKey}, lockValue).Int64()
		if result == 1 {
			fmt.Println("✓ 释放锁成功（原子操作）")
		} else {
			fmt.Println("✗ 释放锁失败（值不匹配，可能已被其他进程持有）")
		}
	} else {
		fmt.Println("✗ 获取锁失败")
	}
}

// 限流器示例
func rateLimiterDemo(rdb *redis.Client) {
	fmt.Println("\n=== 限流器（滑动窗口） ===")

	// Lua 脚本：滑动窗口限流
	rateLimitScript := redis.NewScript(`
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		-- 清理过期记录
		redis.call("ZREMRANGEBYSCORE", key, 0, now - window * 1000)

		-- 获取当前窗口内的请求数
		local count = redis.call("ZCARD", key)

		if count < limit then
			-- 添加新请求
			redis.call("ZADD", key, now, now .. "-" .. math.random())
			redis.call("PEXPIRE", key, window * 1000)
			return 1
		end

		return 0
	`)

	key := "ratelimit:user:123"
	limit := int64(5)         // 限制 5 次
	window := int64(10)       // 10 秒窗口

	// 模拟多次请求
	for i := 1; i <= 7; i++ {
		now := time.Now().UnixMilli()
		result, _ := rateLimitScript.Run(ctx, rdb, []string{key}, limit, window, now).Int64()

		if result == 1 {
			fmt.Printf("✓ 请求 %d: 允许\n", i)
		} else {
			fmt.Printf("✗ 请求 %d: 限流\n", i)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// 清理
	rdb.Del(ctx, key)
}

// 消息队列示例（使用 List）
func messageQueueDemo(rdb *redis.Client) {
	fmt.Println("\n=== 消息队列（List） ===")

	queueKey := "queue:tasks"

	// 生产者：入队
	for i := 1; i <= 3; i++ {
		task := fmt.Sprintf("task_%d", i)
		rdb.RPush(ctx, queueKey, task)
		fmt.Printf("✓ 入队: %s\n", task)
	}

	// 查看队列长度
	length, _ := rdb.LLen(ctx, queueKey).Result()
	fmt.Printf("✓ 队列长度: %d\n", length)

	// 消费者：出队
	for i := 1; i <= 3; i++ {
		// BRPop 阻塞弹出（带超时）
		result, err := rdb.BRPop(ctx, 2*time.Second, queueKey).Result()
		if err == redis.Nil {
			fmt.Println("✗ 队列为空，超时返回")
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("✓ 出队: %s\n", result[1])
	}

	fmt.Println("✓ 消息队列处理完成")
}

// 延迟队列示例（使用 Sorted Set）
func delayedQueueDemo() {
	fmt.Println("\n=== 延迟队列（Sorted Set） ===")
	// 这个函数需要单独演示，因为它需要长时间运行
	// 参考 SKILL.md 中的完整实现
}