package commands

import (
	"context"
	"io"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type AWDChallengeCommandFacade struct {
	core    *AWDChallengeService
	imports *AWDChallengeImportService
}

func NewAWDChallengeCommandFacade(
	db *gorm.DB,
	repo challengeports.AWDChallengeCommandRepository,
) *AWDChallengeCommandFacade {
	return &AWDChallengeCommandFacade{
		core:    NewAWDChallengeService(repo),
		imports: NewAWDChallengeImportService(db, repo),
	}
}

func (s *AWDChallengeCommandFacade) CreateChallenge(
	ctx context.Context,
	actorUserID int64,
	req *dto.CreateAWDChallengeReq,
) (*dto.AWDChallengeResp, error) {
	return s.core.CreateChallenge(ctx, actorUserID, req)
}

func (s *AWDChallengeCommandFacade) UpdateChallenge(
	ctx context.Context,
	id int64,
	req *dto.UpdateAWDChallengeReq,
) (*dto.AWDChallengeResp, error) {
	return s.core.UpdateChallenge(ctx, id, req)
}

func (s *AWDChallengeCommandFacade) DeleteChallenge(ctx context.Context, id int64) error {
	return s.core.DeleteChallenge(ctx, id)
}

func (s *AWDChallengeCommandFacade) PreviewImport(
	ctx context.Context,
	actorUserID int64,
	fileName string,
	reader io.Reader,
) (*dto.AWDChallengeImportPreviewResp, error) {
	return s.imports.PreviewImport(ctx, actorUserID, fileName, reader)
}

func (s *AWDChallengeCommandFacade) ListImports(ctx context.Context, actorUserID int64) ([]dto.AWDChallengeImportPreviewResp, error) {
	return s.imports.ListImports(ctx, actorUserID)
}

func (s *AWDChallengeCommandFacade) GetImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*dto.AWDChallengeImportPreviewResp, error) {
	return s.imports.GetImport(ctx, actorUserID, id)
}

func (s *AWDChallengeCommandFacade) CommitImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*dto.AWDChallengeResp, error) {
	return s.imports.CommitImport(ctx, actorUserID, id)
}
