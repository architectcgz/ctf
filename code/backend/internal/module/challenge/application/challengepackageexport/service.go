package challengepackageexport

import (
	"context"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

type challengeWriteLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*challengeports.ChallengeWriteModel, error)
}

type ChallengePackageExportService struct {
	challengeRepo         challengeWriteLookupRepository
	topologyRepo          challengeports.ChallengeTopologyReadRepository
	packageRepo           challengeports.ChallengePackageRevisionRepository
	packageExportTxRunner challengeports.ChallengePackageExportTxRunner
	packageStorage        challengeports.ChallengePackageStorage
}

func NewChallengePackageExportService(
	challengeRepo challengeWriteLookupRepository,
	topologyRepo challengeports.ChallengeTopologyReadRepository,
	packageRepo challengeports.ChallengePackageRevisionRepository,
	txRunner challengeports.ChallengePackageExportTxRunner,
	packageStorage challengeports.ChallengePackageStorage,
) *ChallengePackageExportService {
	return &ChallengePackageExportService{
		challengeRepo:         challengeRepo,
		topologyRepo:          topologyRepo,
		packageRepo:           packageRepo,
		packageExportTxRunner: txRunner,
		packageStorage:        packageStorage,
	}
}
