package db

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gonet/utils"
	"os"
	"time"
)

const timeoutPing = 3 * time.Second

func ConnectRedis(uri, pwd string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: uri, Password: pwd, DB: db})
	ctx, cancel := utils.MakeCtx(timeoutPing)
	defer cancel()
	//心跳机制
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		logrus.WithField("role", "connect-redis").Error(pong, err)
	}
	return client
}

func RunScript(client *redis.Client, src string, keys []string, args ...interface{}) *redis.Cmd {
	return redis.NewScript(src).Run(client.Context(), client, keys, args...)
}

func RunScriptFromFile(client *redis.Client, path string, keys []string, args ...interface{}) (*redis.Cmd, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cmd := RunScript(client, string(content), keys, args)
	return cmd, cmd.Err()
}

// 发布
func Publish(client *redis.Client, channel, data string) (err error) {
	err = client.Publish(context.Background(), channel, data)
	return err
}

// 订阅
func Subscribe(client *redis.Client, channel string, doFunc func(string)) {
	sub := client.Subscribe(context.Background(), channel)
	_, err := sub.Receive(context.Background())
	if err != nil {
		return
	}
	ch := sub.Channel()
	for msg := range ch {
		doFunc(msg.Payload)
	}
	return
}
