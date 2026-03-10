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
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("连接失败:", err)
	}
	fmt.Println("✓ Redis 连接成功")

	// 字符串操作
	stringOperations(rdb)

	// Hash 操作
	hashOperations(rdb)

	// List 操作
	listOperations(rdb)

	// Set 操作
	setOperations(rdb)

	// Sorted Set 操作
	sortedSetOperations(rdb)

	// Pipeline 操作
	pipelineOperations(rdb)

	// 发布订阅
	pubsubOperations(rdb)

	fmt.Println("\n✓ 所有示例执行完成")
}

// 字符串操作示例
func stringOperations(rdb *redis.Client) {
	fmt.Println("\n=== 字符串操作 ===")

	// SET
	err := rdb.Set(ctx, "name", "张三", 0).Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ SET name 张三")

	// GET
	name, _ := rdb.Get(ctx, "name").Result()
	fmt.Printf("✓ GET name = %s\n", name)

	// SET with expiration
	rdb.Set(ctx, "temp", "临时值", 5*time.Second)
	fmt.Println("✓ SET temp 临时值 EX 5")

	// GETEX
	temp, _ := rdb.GetEx(ctx, "temp", 10*time.Second).Result()
	fmt.Printf("✓ GETEX temp = %s (TTL 延长到 10 秒)\n", temp)

	// INCR
	rdb.Set(ctx, "counter", 0, 0)
	counter, _ := rdb.Incr(ctx, "counter").Result()
	fmt.Printf("✓ INCR counter = %d\n", counter)

	// MSET/MGET
	rdb.MSet(ctx, map[string]interface{}{
		"key1": "值 1",
		"key2": "值 2",
		"key3": "值 3",
	})
	vals, _ := rdb.MGet(ctx, "key1", "key2", "key3").Result()
	fmt.Printf("✓ MGET = %v\n", vals)
}

// Hash 操作示例
func hashOperations(rdb *redis.Client) {
	fmt.Println("\n=== Hash 操作 ===")

	// HSET
	rdb.HSet(ctx, "user:1", "name", "李四")
	rdb.HSet(ctx, "user:1", "age", 25)
	rdb.HSet(ctx, "user:1", "email", "lisi@example.com")
	fmt.Println("✓ HSET user:1 name 李四 age 25 email lisi@example.com")

	// HGET
	name, _ := rdb.HGet(ctx, "user:1", "name").Result()
	fmt.Printf("✓ HGET user:1 name = %s\n", name)

	// HGETALL
	all, _ := rdb.HGetAll(ctx, "user:1").Result()
	fmt.Printf("✓ HGETALL user:1 = %v\n", all)

	// HMSET
	rdb.HMSet(ctx, "user:2", map[string]interface{}{
		"name":  "王五",
		"age":   30,
		"email": "wangwu@example.com",
	})
	fmt.Println("✓ HMSET user:2")

	// HLEN
	count, _ := rdb.HLen(ctx, "user:1").Result()
	fmt.Printf("✓ HLEN user:1 = %d\n", count)
}

// List 操作示例
func listOperations(rdb *redis.Client) {
	fmt.Println("\n=== List 操作 ===")

	// LPUSH
	rdb.LPush(ctx, "tasks", "任务 3", "任务 2", "任务 1")
	fmt.Println("✓ LPUSH tasks 任务 3 任务 2 任务 1")

	// RPUSH
	rdb.RPush(ctx, "tasks", "任务 4")
	fmt.Println("✓ RPUSH tasks 任务 4")

	// LRANGE
	tasks, _ := rdb.LRange(ctx, "tasks", 0, -1).Result()
	fmt.Printf("✓ LRANGE tasks = %v\n", tasks)

	// LPOP
	task, _ := rdb.LPop(ctx, "tasks").Result()
	fmt.Printf("✓ LPOP = %s\n", task)

	// LLEN
	length, _ := rdb.LLen(ctx, "tasks").Result()
	fmt.Printf("✓ LLEN = %d\n", length)
}

// Set 操作示例
func setOperations(rdb *redis.Client) {
	fmt.Println("\n=== Set 操作 ===")

	// SADD
	rdb.SAdd(ctx, "tags", "go", "redis", "database")
	fmt.Println("✓ SADD tags go redis database")

	// SMEMBERS
	tags, _ := rdb.SMembers(ctx, "tags").Result()
	fmt.Printf("✓ SMEMBERS tags = %v\n", tags)

	// SISMEMBER
	exists, _ := rdb.SIsMember(ctx, "tags", "go").Result()
	fmt.Printf("✓ SISMEMBER tags go = %v\n", exists)

	// SINTER
	rdb.SAdd(ctx, "tags2", "go", "python", "redis")
	inter, _ := rdb.SInter(ctx, "tags", "tags2").Result()
	fmt.Printf("✓ SINTER tags tags2 = %v\n", inter)

	// SUNION
	union, _ := rdb.SUnion(ctx, "tags", "tags2").Result()
	fmt.Printf("✓ SUNION tags tags2 = %v\n", union)
}

// Sorted Set 操作示例
func sortedSetOperations(rdb *redis.Client) {
	fmt.Println("\n=== Sorted Set 操作 ===")

	// ZADD
	rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 100, Member: "玩家 A"})
	rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 200, Member: "玩家 B"})
	rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 150, Member: "玩家 C"})
	fmt.Println("✓ ZADD leaderboard")

	// ZRANGE
	range1, _ := rdb.ZRange(ctx, "leaderboard", 0, -1).Result()
	fmt.Printf("✓ ZRANGE leaderboard = %v\n", range1)

	// ZRANGEBYSCORE
	range2, _ := rdb.ZRangeByScore(ctx, "leaderboard", &redis.ZRangeBy{
		Min: "100",
		Max: "180",
	}).Result()
	fmt.Printf("✓ ZRANGEBYSCORE [100,180] = %v\n", range2)

	// ZCARD
	count, _ := rdb.ZCard(ctx, "leaderboard").Result()
	fmt.Printf("✓ ZCARD = %d\n", count)

	// ZREM
	rdb.ZRem(ctx, "leaderboard", "玩家 A")
	fmt.Println("✓ ZREM 玩家 A")
}

// Pipeline 操作示例
func pipelineOperations(rdb *redis.Client) {
	fmt.Println("\n=== Pipeline 操作 ===")

	// 基础 Pipeline
	pipe := rdb.Pipeline()
	pipe.Set(ctx, "pipe:key1", "值 1", 0)
	pipe.Set(ctx, "pipe:key2", "值 2", 0)
	pipe.Set(ctx, "pipe:key3", "值 3", 0)
	pipe.Get(ctx, "pipe:key1")
	pipe.Get(ctx, "pipe:key2")

	cmds, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Pipeline 执行 %d 个命令\n", len(cmds))

	// 处理结果
	for _, cmd := range cmds {
		fmt.Printf("  - %s: %s\n", cmd.Name(), cmd.String())
	}
}

// 发布订阅示例
func pubsubOperations(rdb *redis.Client) {
	fmt.Println("\n=== Pub/Sub 操作 ===")

	// 创建订阅者
	pubsub := rdb.Subscribe(ctx, "news")
	fmt.Println("✓ Subscribe news")

	// 在后台接收消息
	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			fmt.Printf("✓ 收到消息 - 频道：%s, 消息：%s\n", msg.Channel, msg.Payload)
		}
	}()

	// 等待订阅生效
	time.Sleep(100 * time.Millisecond)

	// 发布消息
	err := rdb.Publish(ctx, "news", "重大新闻！").Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Publish news: 重大新闻！")

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 取消订阅
	pubsub.Close()
	fmt.Println("✓ Unsubscribe")
}
