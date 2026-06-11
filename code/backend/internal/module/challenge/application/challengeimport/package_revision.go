package challengeimport

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/module/challenge/domain"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

func (s *ChallengeImportService) createImportedPackageRevision(
	ctx context.Context,
	store challengeports.ChallengeImportTxStore,
	actorUserID int64,
	challenge *challengeports.ImportedChallenge,
	record *challengeports.ChallengeImportPreviewRecord,
	parsed *domain.ParsedChallengePackage,
) (*challengeentity.ChallengePackageRevision, error) {
	if challenge == nil || parsed == nil || record == nil {
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("缺少题目或题包信息"))
	}
	if s.packageStorage == nil {
		return nil, fmt.Errorf("challenge package storage is not configured")
	}

	revisionNo, err := store.NextChallengePackageRevisionNo(ctx, challenge.ID)
	if err != nil {
		return nil, err
	}
	packageSlug := resolveChallengePackageSlug(challenge.PackageSlug, challenge.ID, parsed.Slug)
	stored, err := s.packageStorage.PersistImportedPackageSource(ctx, challengeports.ChallengeImportedPackageSourceRequest{
		ChallengeID:        challenge.ID,
		RevisionNo:         revisionNo,
		PackageSlug:        packageSlug,
		SourceDir:          parsed.RootDir,
		PreviewArchivePath: record.ArchivePath,
		PreviewArchiveName: record.FileName,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	revision := &challengeentity.ChallengePackageRevision{
		ChallengeID:        challenge.ID,
		RevisionNo:         revisionNo,
		SourceType:         challengeentity.ChallengePackageRevisionSourceImported,
		PackageSlug:        packageSlug,
		ArchivePath:        stored.ArchivePath,
		SourceDir:          stored.SourceDir,
		ManifestSnapshot:   parsed.ManifestRaw,
		TopologySourcePath: resolveTopologySourcePath(parsed.Topology),
		TopologySnapshot:   resolveTopologySnapshot(parsed.Topology),
		CreatedBy:          int64Ptr(actorUserID),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := store.CreateImportedPackageRevision(ctx, revision); err != nil {
		return nil, err
	}
	return revision, nil
}

func resolveChallengePackageSlug(packageSlug *string, challengeID int64, fallback string) string {
	if packageSlug != nil && strings.TrimSpace(*packageSlug) != "" {
		return strings.TrimSpace(*packageSlug)
	}
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}
	if challengeID > 0 {
		return fmt.Sprintf("challenge-%d", challengeID)
	}
	return "challenge-package"
}

func resolveTopologySourcePath(topology *domain.ParsedChallengePackageTopology) string {
	if topology == nil {
		return ""
	}
	return strings.TrimSpace(topology.Source)
}

func resolveTopologySnapshot(topology *domain.ParsedChallengePackageTopology) string {
	if topology == nil {
		return ""
	}
	return topology.Raw
}
