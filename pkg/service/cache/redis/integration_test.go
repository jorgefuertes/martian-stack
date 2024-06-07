package redis_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/service/logger"

	redis "git.martianoids.com/martianoids/martian-stack/pkg/service/cache/redis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	goroutineLimit            = 100
	goroutineReadDelayLimitMs = 5
	runLongSec                = 2
	host                      = "127.0.0.1"
	port                      = 6379
	db                        = 2
)

type counter struct {
	n   int
	mux *sync.Mutex
}

func TestCache(t *testing.T) {
	t.Logf("Redis conf: %s:%d@%s:%s/%d", host, port, "", "", db)
	log := logger.New(os.Stderr, logger.TextFormat, logger.LevelDebug)
	c := redis.NewService(log, host, port, "", "", db)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

	t.Cleanup(func() {
		cancel()
		c.Close()
	})

	// set/recover string
	testStr := "This is just a testing str"
	key := "test-str"
	require.NoError(t, c.Set(ctx, key, testStr, 5*time.Second))
	s, err := c.GetString(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, testStr, s)

	rnd := rand.New(rand.NewSource(time.Now().UnixMilli()))

	// set/recover int
	testInt := rnd.Int()
	key = "test-int"
	require.NoError(t, c.Set(ctx, key, testInt, 5*time.Second))
	n, err := c.GetInt(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, testInt, n)

	// set/recover float
	testFloat := rnd.Float64()
	key = "test-float"
	require.NoError(t, c.Set(ctx, key, testFloat, 5*time.Second))
	// check existence
	require.True(t, c.Exists(ctx, key))
	// recover
	f, err := c.GetFloat(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, testFloat, f)
	// delete
	require.NoError(t, c.Delete(ctx, key))
	require.False(t, c.Exists(ctx, key))

	// set and check expiration
	require.NoError(t, c.Set(ctx, key, testFloat, 250*time.Millisecond))
	time.Sleep(500 * time.Millisecond)
	require.False(t, c.Exists(ctx, key))

	ct := new(counter)
	ct.mux = new(sync.Mutex)
	routineCtx, routineCancel := context.WithCancel(context.Background())
	t.Logf("Launching %d goroutines", goroutineLimit)
	wg := new(sync.WaitGroup)
	for i := 0; i < goroutineLimit; i++ {
		wg.Add(1)
		go readAndWrite(t, routineCtx, i, ct, c, wg)
	}
	t.Logf("Letting them run for %d seconds", runLongSec)
	time.Sleep(time.Second * runLongSec)
	routineCancel()
	wg.Wait()
	t.Logf("RW Operation counter: %d", ct.n)
}

func readAndWrite(t *testing.T, ctx context.Context, i int, ct *counter, c *redis.Service, wg *sync.WaitGroup) {
	key := fmt.Sprintf("test-goroutine-%d", i)
	n := 0
	for {
		select {
		case <-ctx.Done():
			ct.mux.Lock()
			ct.n += n
			ct.mux.Unlock()
			wg.Done()
			return
		default:
			cacheCtx, cacheCancel := context.WithTimeout(context.Background(), time.Second*1)
			testInt := rand.Int()
			require.NoError(t, c.Set(cacheCtx, key, testInt, 5*time.Second))
			time.Sleep(time.Duration(rand.Intn(goroutineReadDelayLimitMs)) * time.Millisecond)
			num, err := c.GetInt(cacheCtx, key)
			require.NoError(t, err)
			assert.Equal(t, testInt, num)
			assert.NoError(t, c.Delete(cacheCtx, key))
			n++
			cacheCancel()
		}
	}
}
