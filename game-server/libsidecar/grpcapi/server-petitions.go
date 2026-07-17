package grpcapi

import (
	"context"

	"github.com/walkline/ToCloud9/game-server/libsidecar/queue"
	"github.com/walkline/ToCloud9/gen/worldserver/pb"
)

func (w *WorldServerGRPCAPI) CanTurnInGuildPetition(ctx context.Context, request *pb.CanTurnInGuildPetitionRequest) (*pb.CanTurnInGuildPetitionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	type respType struct {
		resp *GuildPetitionCheckResult
		err  error
	}
	var resp respType

	respChan := make(chan respType, 1)

	w.readQueue.Push(queue.HandlerFunc(func() {
		r, err := w.bindings.CanTurnInGuildPetition(request.PlayerGUID, request.PetitionItemGUID)
		respChan <- respType{
			resp: r,
			err:  err,
		}
		close(respChan)
	}))
	select {
	case <-ctx.Done():
		return nil, ErrTimeout
	case resp = <-respChan:
	}

	if resp.err != nil {
		return nil, resp.err
	}

	var status pb.CanTurnInGuildPetitionResponse_Status
	switch resp.resp.Status {
	case GuildPetitionCheckStatusOk:
		status = pb.CanTurnInGuildPetitionResponse_Ok
	case GuildPetitionCheckStatusPlayerNotFound:
		status = pb.CanTurnInGuildPetitionResponse_PlayerNotFound
	case GuildPetitionCheckStatusPetitionNotFound:
		status = pb.CanTurnInGuildPetitionResponse_PetitionNotFound
	case GuildPetitionCheckStatusNotPetitionOwner:
		status = pb.CanTurnInGuildPetitionResponse_NotPetitionOwner
	case GuildPetitionCheckStatusNotGuildPetition:
		status = pb.CanTurnInGuildPetitionResponse_NotGuildPetition
	case GuildPetitionCheckStatusAlreadyInGuild:
		status = pb.CanTurnInGuildPetitionResponse_AlreadyInGuild
	case GuildPetitionCheckStatusNeedMoreSignatures:
		status = pb.CanTurnInGuildPetitionResponse_NeedMoreSignatures
	default:
		status = pb.CanTurnInGuildPetitionResponse_PetitionNotFound
	}

	return &pb.CanTurnInGuildPetitionResponse{
		Api:            LibVer,
		Status:         status,
		GuildName:      resp.resp.GuildName,
		SignatoryGUIDs: resp.resp.SignatoryGUIDs,
	}, nil
}
