package app

import (
	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type userSelfRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	assessment    *composition.AssessmentModule
	challenge     *composition.ChallengeModule
	contest       *composition.ContestModule
	instance      *composition.InstanceModule
	practice      *composition.PracticeModule
}

func registerUserSelfRoutes(apiV1, protected *gin.RouterGroup, deps userSelfRouteDeps) {
	registerUserContestRoutes(apiV1, protected, userContestRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		contest:       deps.contest,
		instance:      deps.instance,
		practice:      deps.practice,
	})
	registerUserPracticeRoutes(apiV1, protected, userPracticeRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		challenge:     deps.challenge,
		instance:      deps.instance,
		practice:      deps.practice,
	})
	registerUserSelfServiceRoutes(protected, userSelfServiceRouteDeps{
		assessment: deps.assessment,
		practice:   deps.practice,
	})
}
