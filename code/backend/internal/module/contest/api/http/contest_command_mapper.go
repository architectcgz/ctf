package http

import (
	"ctf-platform/internal/dto"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
)

func createContestInputFromDTO(req *dto.CreateContestReq) contestcmd.CreateContestInput {
	if req == nil {
		return contestcmd.CreateContestInput{}
	}
	return contestRequestMapper.ToCreateContestInput(*req)
}

func updateContestInputFromDTO(req *dto.UpdateContestReq) contestcmd.UpdateContestInput {
	if req == nil {
		return contestcmd.UpdateContestInput{}
	}
	return contestRequestMapper.ToUpdateContestInput(*req)
}

func createAnnouncementInputFromDTO(req *dto.CreateContestAnnouncementReq) contestcmd.CreateAnnouncementInput {
	if req == nil {
		return contestcmd.CreateAnnouncementInput{}
	}
	return contestRequestMapper.ToCreateAnnouncementInput(*req)
}

func reviewRegistrationInputFromDTO(req *dto.ReviewContestRegistrationReq) contestcmd.ReviewRegistrationInput {
	if req == nil {
		return contestcmd.ReviewRegistrationInput{}
	}
	return contestRequestMapper.ToReviewRegistrationInput(*req)
}

func createTeamInputFromDTO(req *dto.CreateTeamReq) contestcmd.CreateTeamInput {
	if req == nil {
		return contestcmd.CreateTeamInput{}
	}
	return contestRequestMapper.ToCreateTeamInput(*req)
}

func addContestChallengeInputFromDTO(req *dto.AddContestChallengeReq) contestcmd.AddContestChallengeInput {
	if req == nil {
		return contestcmd.AddContestChallengeInput{}
	}
	return contestRequestMapper.ToAddContestChallengeInput(*req)
}

func updateContestChallengeInputFromDTO(req *dto.UpdateContestChallengeReq) contestcmd.UpdateContestChallengeInput {
	if req == nil {
		return contestcmd.UpdateContestChallengeInput{}
	}
	return contestRequestMapper.ToUpdateContestChallengeInput(*req)
}

func createAWDRoundInputFromDTO(req *dto.CreateAWDRoundReq) contestcmd.CreateAWDRoundInput {
	if req == nil {
		return contestcmd.CreateAWDRoundInput{}
	}
	return contestRequestMapper.ToCreateAWDRoundInput(*req)
}

func upsertServiceCheckInputFromDTO(req *dto.UpsertAWDServiceCheckReq) contestcmd.UpsertServiceCheckInput {
	if req == nil {
		return contestcmd.UpsertServiceCheckInput{}
	}
	return contestRequestMapper.ToUpsertServiceCheckInput(*req)
}

func runCurrentRoundChecksInputFromDTO(req *dto.RunCurrentAWDCheckerReq) contestcmd.RunCurrentRoundChecksInput {
	if req == nil {
		return contestcmd.RunCurrentRoundChecksInput{}
	}
	return contestRequestMapper.ToRunCurrentRoundChecksInput(*req)
}

func createAttackLogInputFromDTO(req *dto.CreateAWDAttackLogReq) contestcmd.CreateAttackLogInput {
	if req == nil {
		return contestcmd.CreateAttackLogInput{}
	}
	return contestRequestMapper.ToCreateAttackLogInput(*req)
}

func submitAttackInputFromDTO(req *dto.SubmitAWDAttackReq) contestcmd.SubmitAttackInput {
	if req == nil {
		return contestcmd.SubmitAttackInput{}
	}
	return contestRequestMapper.ToSubmitAttackInput(*req)
}

func previewCheckerInputFromDTO(req *dto.PreviewAWDCheckerReq) contestcmd.PreviewCheckerInput {
	if req == nil {
		return contestcmd.PreviewCheckerInput{}
	}
	return contestRequestMapper.ToPreviewCheckerInput(*req)
}

func createContestAWDServiceInputFromDTO(req *dto.CreateContestAWDServiceReq) contestcmd.CreateContestAWDServiceInput {
	if req == nil {
		return contestcmd.CreateContestAWDServiceInput{}
	}
	return contestRequestMapper.ToCreateContestAWDServiceInput(*req)
}

func updateContestAWDServiceInputFromDTO(req *dto.UpdateContestAWDServiceReq) contestcmd.UpdateContestAWDServiceInput {
	if req == nil {
		return contestcmd.UpdateContestAWDServiceInput{}
	}
	return contestRequestMapper.ToUpdateContestAWDServiceInput(*req)
}
