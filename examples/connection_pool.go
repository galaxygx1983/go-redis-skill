package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	// 创建连接池
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		PoolSize: 100, // 连接池大小
	})
	defer rdb.Close()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("连接失败:", err)
	}
	fmt.Println("✓ Redis 连接池已创建")

	// 演示连接池使用
	connectionPoolDemo(rdb)

	// 演示并发访问
	concurrentAccessDemo(rdb)

	// 演示超时处理
	timeoutDemo(rdb)

	fmt.Println("\n✓ 所有连接池示例执行完成")
}

// 连接池使用示例
func connectionPoolDemo(rdb *redis.Client) {
	fmt.Println("\n=== 连接池使用 ===")

	// 获取连接池统计信息
	stats := rdb.PoolStats()
	fmt.Printf("✓ 连接池统计:\n")
	fmt.Printf("  - 总连接数：%d\n", stats.TotalConns)
	fmt.Printf("  - 空闲连接数：%d\n", stats.IdleConns)
	fmt.Printf("  - 活跃连接数：%d\n", stats.StaleConns)

	// 模拟多次操作
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("pool:key%d", i)
		rdb.Set(ctx, key, i, 0)
	}

	stats = rdb.PoolStats()
	fmt.Printf("✓ 执行 10 次 SET 后:\n")
	fmt.Printf("  - 总连接数：%d\n", stats.TotalConns)
	fmt.Printf("  - 空闲连接数：%d\n", stats.IdleConns)
}

// 并发访问示例
func concurrentAccessDemo(rdb *redis.Client) {
	fmt.Println("\n=== 并发访问 ===")

	var wg sync.WaitGroup
	workers := 10
	requestsPerWorker := 100

	// 启动多个 worker
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for i := 0; i < requestsPerWorker; i++ {
				key := fmt.Sprintf("concurrent:worker%d:key%d", workerID, i)
				value := fmt.Sprintf("value%d", i)

				// SET
				if err := rdb.Set(ctx, key, value, 0).Err(); err != nil {
					log.Printf("Worker %d SET error: %v", workerID, err)
					continue
				}

				// GET
				val, err := rdb.Get(ctx, key).Result()
				if err != nil {
					log.Printf("Worker %d GET error: %v", workerID, err)
					continue
				}

				if val != value {
					log.Printf("Worker %d: value mismatch", workerID)
				}
			}
		}(w)
	}

	// 等待所有 worker 完成
	wg.Wait()

	stats := rdb.PoolStats()
	fmt.Printf("✓ 并发访问完成 (%d workers × %d requests)\n", workers, requestsPerWorker)
	fmt.Printf("  - 总连接数：%d\n", stats.TotalConns)
	fmt.Printf("  - 空闲连接数：%d\n", stats.IdleConns)
}

// 超时处理示例
func timeoutDemo(rdb *redis.Client) {
	fmt.Println("\n=== 超时处理 ===")

	// 正常请求
	ctx1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()

	start := time.Now()
	err := rdb.Ping(ctx1).Err()
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("✗ Ping 失败：%v\n", err)
	} else {
		fmt.Printf("✓ Ping 成功 (耗时：%v)\n", duration)
	}

	// 模拟超时（使用不存在的 Redis 地址）
	timeoutClient := redis.NewClient(&redis.Options{
		Addr:        "localhost:9999", // 不存在的端口
		DialTimeout: 1 * time.Second,
	})
	defer timeoutClient.Close()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()

	start = time.Now()
	err = timeoutClient.Ping(ctx2).Err()
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("✓ 预期超时 (耗时：%v): %v\n", duration, err)
	}
}

// 连接池配置示例
func connectionPoolConfig() {
	fmt.Println("\n=== 连接池配置 ===")

	// 推荐配置
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",

		// 连接池大小
		PoolSize:     100,             // 最大连接数
		PoolTimeout:  4 * time.Second, // 连接池超时
		MinIdleConns: 10,              // 最小空闲连接

		// 连接超时
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	defer rdb.Close()

	stats := rdb.PoolStats()
	fmt.Printf("✓ 连接池配置完成:\n")
	fmt.Printf("  - PoolSize: 100\n")
	fmt.Printf("  - MinIdleConns: 10\n")
	fmt.Printf("  - 当前总连接数：%d\n", stats.TotalConns)
}
