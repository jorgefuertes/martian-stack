package memory_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/jorgefuertes/martian-stack/pkg/service/cache/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	goroutineLimit            = 10
	goroutineReadDelayLimitMs = 1
	runLongSec                = 2
)

type counter struct {
	n   int
	mux *sync.Mutex
}

type testObject struct {
	Name  string  `json:"name"`
	Age   int     `json:"age"`
	Money float64 `json:"money"`
}

func TestCache(t *testing.T) {
	c := memory.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

	t.Cleanup(func() {
		cancel()
		c.Close()
	})

	t.Run("string", func(t *testing.T) {
		testStr := "This is just a testing str"
		key := "test-str"
		require.NoError(t, c.Set(ctx, key, testStr, 5*time.Second))
		s, err := c.GetString(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, testStr, s)
	})

	t.Run("int", func(t *testing.T) {
		rnd := rand.New(rand.NewSource(time.Now().UnixMilli()))
		testInt := rnd.Int()
		key := "test-int"
		require.NoError(t, c.Set(ctx, key, testInt, 5*time.Second))
		n, err := c.GetInt(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, testInt, n)
	})

	t.Run("float", func(t *testing.T) {
		rnd := rand.New(rand.NewSource(time.Now().UnixMilli()))
		testFloat := rnd.Float64()
		key := "test-float"
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
		key += "_2"
		require.NoError(t, c.Set(ctx, key, testFloat, 250*time.Millisecond))
		require.True(t, shouldExpire(ctx, c, key, 10*time.Second), "Not expired: %s", key)
	})

	t.Run("object", func(t *testing.T) {
		// set/recover object
		testObj := &testObject{Name: "John", Age: 30, Money: 123.4}
		key := "test-obj"
		require.NoError(t, c.Set(ctx, key, testObj, 5*time.Second))
		require.True(t, c.Exists(ctx, key))
		obj := new(testObject)
		require.NoError(t, c.Get(ctx, key, obj))
		require.Equal(t, testObj, obj)
		require.NoError(t, c.Delete(ctx, key))
		require.False(t, c.Exists(ctx, key))
	})

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

func readAndWrite(t *testing.T, ctx context.Context, i int, ct *counter, c *memory.Service, wg *sync.WaitGroup) {
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

func shouldExpire(ctx context.Context, c *memory.Service, key string, limit time.Duration) bool {
	start := time.Now()
	for start.Add(limit).After(time.Now()) {
		time.Sleep(50 * time.Millisecond)
		if !c.Exists(ctx, key) {
			return true
		}
	}

	return false
}
