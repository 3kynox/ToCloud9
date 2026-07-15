package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_guildNamesCached_GuildNameByID(t *testing.T) {
	const (
		realmID = uint32(1)
		guildID = uint32(64)
	)

	t.Run("caches source result", func(t *testing.T) {
		calls := 0
		cache := newGuildNamesCached(func(_ context.Context, _ uint32, _ uint32) (string, error) {
			calls++
			return "Moebius", nil
		}, time.Minute, time.Now)

		for i := 0; i < 3; i++ {
			name, err := cache.GuildNameByID(context.Background(), realmID, guildID)
			assert.NoError(t, err)
			assert.Equal(t, "Moebius", name)
		}
		assert.Equal(t, 1, calls)
	})

	t.Run("expired entry re-fetches", func(t *testing.T) {
		calls := 0
		current := time.Now()
		cache := newGuildNamesCached(func(_ context.Context, _ uint32, _ uint32) (string, error) {
			calls++
			return "Moebius", nil
		}, time.Minute, func() time.Time { return current })

		_, err := cache.GuildNameByID(context.Background(), realmID, guildID)
		assert.NoError(t, err)

		current = current.Add(2 * time.Minute)
		_, err = cache.GuildNameByID(context.Background(), realmID, guildID)
		assert.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("caches missing guild as empty name", func(t *testing.T) {
		calls := 0
		cache := newGuildNamesCached(func(_ context.Context, _ uint32, _ uint32) (string, error) {
			calls++
			return "", nil
		}, time.Minute, time.Now)

		for i := 0; i < 2; i++ {
			name, err := cache.GuildNameByID(context.Background(), realmID, guildID)
			assert.NoError(t, err)
			assert.Equal(t, "", name)
		}
		assert.Equal(t, 1, calls)
	})

	t.Run("zero guild id short-circuits", func(t *testing.T) {
		cache := newGuildNamesCached(func(_ context.Context, _ uint32, _ uint32) (string, error) {
			t.Fatal("source should not be called")
			return "", nil
		}, time.Minute, time.Now)

		name, err := cache.GuildNameByID(context.Background(), realmID, 0)
		assert.NoError(t, err)
		assert.Equal(t, "", name)
	})

	t.Run("source error is not cached", func(t *testing.T) {
		calls := 0
		cache := newGuildNamesCached(func(_ context.Context, _ uint32, _ uint32) (string, error) {
			calls++
			return "", assert.AnError
		}, time.Minute, time.Now)

		_, err := cache.GuildNameByID(context.Background(), realmID, guildID)
		assert.Error(t, err)
		_, err = cache.GuildNameByID(context.Background(), realmID, guildID)
		assert.Error(t, err)
		assert.Equal(t, 2, calls)
	})
}
