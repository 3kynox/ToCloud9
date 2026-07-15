package repo

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	shrepo "github.com/walkline/ToCloud9/shared/repo"
)

// guildNamesCacheTTL bounds staleness of resolved guild names. Guild ids can be
// reused after a disband (ids are allocated as MAX(guildid)+1), so entries expire.
const guildNamesCacheTTL = time.Minute

// GuildNameResolver resolves guild names by guild id.
type GuildNameResolver interface {
	// GuildNameByID returns the guild name or empty string if the guild doesn't exist.
	GuildNameByID(ctx context.Context, realmID uint32, guildID uint32) (string, error)
}

// guildNameSource fetches a guild name from the source of truth.
// Returns empty string when the guild doesn't exist.
type guildNameSource func(ctx context.Context, realmID uint32, guildID uint32) (string, error)

type guildNameEntry struct {
	name      string
	fetchedAt time.Time
}

// guildNamesCached is a GuildNameResolver with a TTL in-mem cache on top of a source.
// Missing guilds are cached as empty names to avoid hammering the source.
type guildNamesCached struct {
	source guildNameSource
	ttl    time.Duration
	now    func() time.Time

	mu    sync.RWMutex
	cache map[uint32]map[uint32]guildNameEntry
}

func newGuildNamesCached(source guildNameSource, ttl time.Duration, now func() time.Time) *guildNamesCached {
	return &guildNamesCached{
		source: source,
		ttl:    ttl,
		now:    now,
		cache:  map[uint32]map[uint32]guildNameEntry{},
	}
}

// NewGuildNamesMySQL returns GuildNameResolver backed by the characters DB.
func NewGuildNamesMySQL(db shrepo.CharactersDB) GuildNameResolver {
	return newGuildNamesCached(func(ctx context.Context, realmID uint32, guildID uint32) (string, error) {
		var name string
		err := db.DBByRealm(realmID).QueryRowContext(ctx, "SELECT name FROM guild WHERE guildid = ?", guildID).Scan(&name)
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return name, nil
	}, guildNamesCacheTTL, time.Now)
}

// GuildNameByID returns the guild name or empty string if the guild doesn't exist.
func (g *guildNamesCached) GuildNameByID(ctx context.Context, realmID uint32, guildID uint32) (string, error) {
	if guildID == 0 {
		return "", nil
	}

	g.mu.RLock()
	entry, found := g.cache[realmID][guildID]
	g.mu.RUnlock()
	if found && g.now().Sub(entry.fetchedAt) < g.ttl {
		return entry.name, nil
	}

	name, err := g.source(ctx, realmID, guildID)
	if err != nil {
		return "", err
	}

	g.mu.Lock()
	if g.cache[realmID] == nil {
		g.cache[realmID] = map[uint32]guildNameEntry{}
	}
	g.cache[realmID][guildID] = guildNameEntry{name: name, fetchedAt: g.now()}
	g.mu.Unlock()

	return name, nil
}
