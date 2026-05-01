package http

import (
	"ctf-platform/internal/dto"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
)

func createContestInputFromDTO(req *dto.CreateContestReq) contestcmd.CreateContestInput {
	if req == nil {
		return contestcmd.CreateContestInput{}
	}
	return contestcmd.CreateContestInput{
		Title:       req.Title,
		Description: req.Description,
		Mode:        req.Mode,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}
}

func updateContestInputFromDTO(req *dto.UpdateContestReq) contestcmd.UpdateContestInput {
	if req == nil {
		return contestcmd.UpdateContestInput{}
	}
	return contestcmd.UpdateContestInput{
		Title:          req.Title,
		Description:    req.Description,
		Mode:           req.Mode,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Status:         req.Status,
		ForceOverride:  req.ForceOverride,
		OverrideReason: req.OverrideReason,
	}
}

func createAnnouncementInputFromDTO(req *dto.CreateContestAnnouncementReq) contestcmd.CreateAnnouncementInput {
	if req == nil {
		return contestcmd.CreateAnnouncementInput{}
	}
	return contestcmd.CreateAnnouncementInput{
		Title:   req.Title,
		Content: req.Content,
	}
}

func reviewRegistrationInputFromDTO(req *dto.ReviewContestRegistrationReq) contestcmd.ReviewRegistrationInput {
	if req == nil {
		return contestcmd.ReviewRegistrationInput{}
	}
	return contestcmd.ReviewRegistrationInput{
		Status: req.Status,
	}
}

func createTeamInputFromDTO(req *dto.CreateTeamReq) contestcmd.CreateTeamInput {
	if req == nil {
		return contestcmd.CreateTeamInput{}
	}
	return contestcmd.CreateTeamInput{
		Name:       req.Name,
		MaxMembers: req.MaxMembers,
	}
}

func addContestChallengeInputFromDTO(req *dto.AddContestChallengeReq) contestcmd.AddContestChallengeInput {
	if req == nil {
		return contestcmd.AddContestChallengeInput{}
	}
	return contestcmd.AddContestChallengeInput{
		ChallengeID: req.ChallengeID,
		Points:      req.Points,
		Order:       req.Order,
		IsVisible:   req.IsVisible,
	}
}

func updateContestChallengeInputFromDTO(req *dto.UpdateContestChallengeReq) contestcmd.UpdateContestChallengeInput {
	if req == nil {
		return contestcmd.UpdateContestChallengeInput{}
	}
	return contestcmd.UpdateContestChallengeInput{
		Points:    req.Points,
		Order:     req.Order,
		IsVisible: req.IsVisible,
	}
}

func createAWDRoundInputFromDTO(req *dto.CreateAWDRoundReq) contestcmd.CreateAWDRoundInput {
	if req == nil {
		return contestcmd.CreateAWDRoundInput{}
	}
	return contestcmd.CreateAWDRoundInput{
		RoundNumber:    req.RoundNumber,
		Status:         req.Status,
		AttackScore:    req.AttackScore,
		DefenseScore:   req.DefenseScore,
		ForceOverride:  req.ForceOverride,
		OverrideReason: req.OverrideReason,
	}
}

func upsertServiceCheckInputFromDTO(req *dto.UpsertAWDServiceCheckReq) contestcmd.UpsertServiceCheckInput {
	if req == nil {
		return contestcmd.UpsertServiceCheckInput{}
	}
	return contestcmd.UpsertServiceCheckInput{
		TeamID:        req.TeamID,
		ServiceID:     req.ServiceID,
		ServiceStatus: req.ServiceStatus,
		CheckResult:   req.CheckResult,
	}
}

func runCurrentRoundChecksInputFromDTO(req *dto.RunCurrentAWDCheckerReq) contestcmd.RunCurrentRoundChecksInput {
	if req == nil {
		return contestcmd.RunCurrentRoundChecksInput{}
	}
	return contestcmd.RunCurrentRoundChecksInput{
		ForceOverride:  req.ForceOverride,
		OverrideReason: req.OverrideReason,
	}
}

func createAttackLogInputFromDTO(req *dto.CreateAWDAttackLogReq) contestcmd.CreateAttackLogInput {
	if req == nil {
		return contestcmd.CreateAttackLogInput{}
	}
	return contestcmd.CreateAttackLogInput{
		AttackerTeamID: req.AttackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ServiceID:      req.ServiceID,
		AttackType:     req.AttackType,
		SubmittedFlag:  req.SubmittedFlag,
		IsSuccess:      req.IsSuccess,
	}
}
