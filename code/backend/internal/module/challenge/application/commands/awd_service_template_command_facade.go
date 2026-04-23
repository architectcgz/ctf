package commands

import (
	"context"
	"io"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type AWDServiceTemplateCommandFacade struct {
	core    *AWDServiceTemplateService
	imports *AWDServiceTemplateImportService
}

func NewAWDServiceTemplateCommandFacade(
	db *gorm.DB,
	repo challengeports.AWDServiceTemplateCommandRepository,
) *AWDServiceTemplateCommandFacade {
	return &AWDServiceTemplateCommandFacade{
		core:    NewAWDServiceTemplateService(repo),
		imports: NewAWDServiceTemplateImportService(db, repo),
	}
}

func (s *AWDServiceTemplateCommandFacade) CreateTemplate(
	actorUserID int64,
	req *dto.CreateAWDServiceTemplateReq,
) (*dto.AWDServiceTemplateResp, error) {
	return s.core.CreateTemplate(actorUserID, req)
}

func (s *AWDServiceTemplateCommandFacade) CreateTemplateWithContext(
	ctx context.Context,
	actorUserID int64,
	req *dto.CreateAWDServiceTemplateReq,
) (*dto.AWDServiceTemplateResp, error) {
	return s.core.CreateTemplateWithContext(ctx, actorUserID, req)
}

func (s *AWDServiceTemplateCommandFacade) UpdateTemplate(
	id int64,
	req *dto.UpdateAWDServiceTemplateReq,
) (*dto.AWDServiceTemplateResp, error) {
	return s.core.UpdateTemplate(id, req)
}

func (s *AWDServiceTemplateCommandFacade) UpdateTemplateWithContext(
	ctx context.Context,
	id int64,
	req *dto.UpdateAWDServiceTemplateReq,
) (*dto.AWDServiceTemplateResp, error) {
	return s.core.UpdateTemplateWithContext(ctx, id, req)
}

func (s *AWDServiceTemplateCommandFacade) DeleteTemplate(id int64) error {
	return s.core.DeleteTemplate(id)
}

func (s *AWDServiceTemplateCommandFacade) DeleteTemplateWithContext(ctx context.Context, id int64) error {
	return s.core.DeleteTemplateWithContext(ctx, id)
}

func (s *AWDServiceTemplateCommandFacade) PreviewImport(
	ctx context.Context,
	actorUserID int64,
	fileName string,
	reader io.Reader,
) (*dto.AWDServiceTemplateImportPreviewResp, error) {
	return s.imports.PreviewImport(ctx, actorUserID, fileName, reader)
}

func (s *AWDServiceTemplateCommandFacade) ListImports(actorUserID int64) ([]dto.AWDServiceTemplateImportPreviewResp, error) {
	return s.imports.ListImports(actorUserID)
}

func (s *AWDServiceTemplateCommandFacade) GetImport(
	actorUserID int64,
	id string,
) (*dto.AWDServiceTemplateImportPreviewResp, error) {
	return s.imports.GetImport(actorUserID, id)
}

func (s *AWDServiceTemplateCommandFacade) CommitImport(
	ctx context.Context,
	actorUserID int64,
	id string,
) (*dto.AWDServiceTemplateResp, error) {
	return s.imports.CommitImport(ctx, actorUserID, id)
}
