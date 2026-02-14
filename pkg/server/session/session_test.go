package session_test

import (
	"context"
	"testing"

	"github.com/jorgefuertes/martian-stack/pkg/server/session"
	"github.com/jorgefuertes/martian-stack/pkg/service/cache"
	"github.com/jorgefuertes/martian-stack/pkg/service/cache/memory"
	"github.com/jorgefuertes/martian-stack/pkg/service/cache/redis"
	"github.com/jorgefuertes/martian-stack/pkg/service/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	cacheSvcs := []struct {
		name string
		svc  cache.Service
	}{
		{name: "memory", svc: memory.New()},
		{
			name: "redis",
			svc:  redis.New(logger.NewNull(), redis.DefaultHost, redis.DefaultPort, "", "", redis.DefaultDB),
		},
	}

	defer func() {
		for _, tc := range cacheSvcs {
			tc.svc.Close()
		}
	}()

	for _, tc := range cacheSvcs {
		t.Run(tc.name, func(t *testing.T) {
			s := session.New()
			assert.NotEmpty(t, s.ID)
			assert.NotNil(t, s.Data())

			// set
			err := s.Data().Set("test-key", "test-value")
			require.NoError(t, err)
			require.True(t, s.Data().IsDirty())
			v := s.Data().GetString("test-key")
			assert.Equal(t, "test-value", v)

			// save
			b, err := s.MarshalJSON()
			require.NoError(t, err)
			err = tc.svc.Set(context.Background(), s.ID, b, 0)
			require.NoError(t, err)

			// recover
			b2, err := tc.svc.GetBytes(context.Background(), s.ID)
			require.NoError(t, err)
			s2 := session.New().WithID(s.ID)
			err = s2.UnmarshalJSON(b2)
			require.NoError(t, err)
			assert.Equal(t, s.ID, s2.ID)
			assert.Equal(t, "test-value", s2.Data().GetString("test-key"))
		})
	}
}
