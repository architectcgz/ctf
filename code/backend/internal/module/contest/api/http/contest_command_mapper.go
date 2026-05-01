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
