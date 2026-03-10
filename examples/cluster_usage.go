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
	// 创建 Cluster 客户端
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:7000",
			"localhost:7001",
			"localhost:7002",
		},
		Password: "", // 如果有密码
	})
	defer rdb.Close()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("集群连接失败:", err)
	}
	fmt.Println("✓ Redis Cluster 连接成功")

	// 演示 Cluster 操作
	clusterOperations(rdb)

	fmt.Println("\n✓ Cluster 示例执行完成")
}

// Cluster 操作示例
func clusterOperations(rdb *redis.ClusterClient) {
	fmt.Println("\n=== Cluster 操作 ===")

	// 基础 SET/GET
	err := rdb.Set(ctx, "user:1:name", "张三", 0).Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ SET user:1:name 张三")

	name, err := rdb.Get(ctx, "user:1:name").Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ GET user:1:name = %s\n", name)

	// 查看键分布在哪个节点
	slot := rdb.SlotForKey("user:1:name")
	fmt.Printf("✓ user:1:name 的 slot = %d\n", slot)

	// 遍历所有分片
	err = rdb.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
		info, err := shard.Info(ctx).Result()
		if err != nil {
			return err
		}
		fmt.Printf("✓ 分片节点：%s\n", shard.String())
		fmt.Printf("  信息：%s\n", info[:100]) // 只显示前 100 字符
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Pipeline 操作
	pipe := rdb.Pipeline()
	pipe.Set(ctx, "cluster:key1", "值 1", 0)
	pipe.Set(ctx, "cluster:key2", "值 2", 0)
	pipe.Get(ctx, "cluster:key1")

	cmds, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Pipeline 执行 %d 个命令\n", len(cmds))
}

// Cluster 配置示例
func clusterConfigExample() {
	fmt.Println("\n=== Cluster 配置 ===")

	// 推荐配置
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:7000",
			"localhost:7001",
			"localhost:7002",
		},

		// 认证
		Password: "password",

		// 路由选项
		RouteByLatency: true, // 按延迟路由
		RouteRandomly:  true, // 随机路由

		// 连接池
		PoolSize:     100,
		MinIdleConns: 10,

		// 超时
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	defer rdb.Close()

	fmt.Println("✓ Cluster 配置完成")
}
