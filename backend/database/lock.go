package database

import (
	"context"
	"errors"
	"time"
)

const lockTTL = 10 * time.Second

var ErrLockFailed = errors.New("获取分布式锁失败，请稍后重试")

// AcquireLock 尝试获取分布式锁，返回释放函数
func AcquireLock(ctx context.Context, key string) (release func(), err error) {
	ok, err := RDB.SetNX(ctx, key, 1, lockTTL).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLockFailed
	}
	return func() {
		RDB.Del(context.Background(), key)
	}, nil
}
