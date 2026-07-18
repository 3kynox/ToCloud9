package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/walkline/ToCloud9/shared/events"
)

func TestMembersStatsCollectorDropsPendingUpdatesOnLogout(t *testing.T) {
	lvl := uint8(4)
	collector := NewMembersStatsCollector(nil, nil, nil, 0)

	err := collector.HandleCharactersUpdates(events.GWEventCharactersUpdatesPayload{
		RealmID: 1,
		Updates: []*events.CharacterUpdate{{ID: 40554, Lvl: &lvl}},
	})
	assert.NoError(t, err)
	assert.Contains(t, collector.pending[1], uint64(40554))

	err = collector.HandleCharacterLoggedOut(events.GWEventCharacterLoggedOutPayload{
		RealmID:  1,
		CharGUID: 40554,
	})
	assert.NoError(t, err)
	assert.NotContains(t, collector.pending[1], uint64(40554))
}

func TestMergeCharacterUpdatePositionAndDeathState(t *testing.T) {
	x, y := float32(12.5), float32(-33.0)
	dead, ghost := true, false

	dst := events.CharacterUpdate{ID: 1}
	mergeCharacterUpdate(&dst, &events.CharacterUpdate{ID: 1, PosX: &x, PosY: &y, IsDead: &dead, IsGhost: &ghost})

	require.NotNil(t, dst.PosX)
	require.NotNil(t, dst.PosY)
	require.NotNil(t, dst.IsDead)
	require.NotNil(t, dst.IsGhost)

	// A later update without those fields keeps the previous values.
	hp := uint32(100)
	mergeCharacterUpdate(&dst, &events.CharacterUpdate{ID: 1, CurHP: &hp})
	require.NotNil(t, dst.PosX)
	require.NotNil(t, dst.IsDead)
}
