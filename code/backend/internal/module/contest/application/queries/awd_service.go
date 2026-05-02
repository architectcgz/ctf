package queries

import contestports "ctf-platform/internal/module/contest/ports"

type awdQueryRepository interface {
	contestports.AWDRoundStore
	contestports.AWDTeamLookup
	contestports.AWDServiceDefinitionQuery
	contestports.AWDReadinessQuery
	contestports.AWDServiceInstanceQuery
	contestports.AWDServiceOperationQuery
	contestports.AWDTeamServiceStore
	contestports.AWDAttackLogStore
	contestports.AWDTrafficEventQuery
}

type AWDService struct {
	repo        awdQueryRepository
	contestRepo contestports.ContestLookupRepository
}

func NewAWDService(repo awdQueryRepository, contestRepo contestports.ContestLookupRepository) *AWDService {
	return &AWDService{
		repo:        repo,
		contestRepo: contestRepo,
	}
}
