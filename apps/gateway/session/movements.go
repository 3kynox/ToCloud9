package session

import (
	"context"
	"time"

	"github.com/walkline/ToCloud9/apps/gateway/packet"

	"github.com/walkline/ToCloud9/shared/events"
)

// posPublishMinInterval throttles group position publishing: movement packets
// arrive ~every 500ms while moving, the party map dot doesn't need that rate.
const posPublishMinInterval = 3 * time.Second

func (s *GameSession) HandleMovement(ctx context.Context, p *packet.Packet) error {
	defer func() {
		if p.Source == packet.SourceGameClient && s.worldSocket != nil {
			s.worldSocket.SendPacket(p)
			return
		}

		if p.Source == packet.SourceWorldServer && s.gameSocket != nil {
			s.gameSocket.SendPacket(p)
			return
		}
	}()

	if p.Source == packet.SourceWorldServer {
		return nil
	}

	r := p.Reader()
	if r.ReadGUID() != s.character.GUID {
		return nil
	}

	_ = r.Uint32() // flags
	_ = r.Uint16() // flags2
	_ = r.Uint32() // time

	s.character.PositionX, s.character.PositionY, s.character.PositionZ, s.character.PositionO = r.Float32(), r.Float32(), r.Float32(), r.Float32()

	s.maybePublishPosition()

	return nil
}

func (s *GameSession) maybePublishPosition() {
	char := s.character
	if char.PositionX == char.lastPublishedPosX && char.PositionY == char.lastPublishedPosY {
		return
	}

	now := time.Now()
	if now.Sub(char.lastPosPublishAt) < posPublishMinInterval {
		return
	}

	char.lastPosPublishAt = now
	char.lastPublishedPosX, char.lastPublishedPosY = char.PositionX, char.PositionY
	posX, posY := char.PositionX, char.PositionY
	s.charsUpdsBarrier.Update(events.CharacterUpdate{ID: char.GUID, PosX: &posX, PosY: &posY})
}
