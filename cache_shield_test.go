package cacheshield

import (
	"context"
	"github.com/avast/retry-go/v4"
	redisv7 "github.com/go-redis/redis/v7"
	"log"
	"sync"
	"testing"
	"time"
)

func TestCacheShield_LoadOrStore(t *testing.T) {
	ctx := context.Background()

	client := redisv7.NewClient(&redisv7.Options{
		Addr:     "127.0.0.1:6379",
		Password: "password",
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		t.Fatal(err)
	}
	cs := NewV7(client)

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result, actual, err := cs.LoadOrStore(ctx, "testKey", callback,
				WithExpiration(10*time.Minute),
				WithLockExpiration(10*time.Second),
				WithRetryOptions(
					retry.Context(ctx),
					retry.Attempts(12),
					retry.Delay(200*time.Millisecond),
					retry.MaxJitter(100*time.Millisecond),
					retry.DelayType(retry.CombineDelay(retry.FixedDelay, retry.RandomDelay)),
					retry.LastErrorOnly(true),
				),
			)
			t.Logf("%d: result: %v actual: %v, error: %v", i, result, actual, err)
		}(i)
	}
	wg.Wait()
}

func callback(ctx context.Context) (string, error) {
	time.Sleep(time.Second)
	log.Println("callback")
	return "testValue", nil
}
