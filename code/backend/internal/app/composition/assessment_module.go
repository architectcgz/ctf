package composition

import (
	assessmentruntime "ctf-platform/internal/module/assessment/runtime"
)

type AssessmentModule = assessmentruntime.Module

func BuildAssessmentModule(root *Root, challenge *ChallengeModule) *AssessmentModule {
	module := assessmentruntime.Build(assessmentruntime.Deps{
		AppContext:    root.Context(),
		Config:        root.Config(),
		Logger:        root.Logger(),
		DB:            root.DB(),
		Cache:         root.Cache(),
		Events:        root.Events,
		ChallengeRepo: challenge.Catalog,
	})
	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewBackgroundJob(job.Name, job.Start, job.Stop))
	}
	return module
}
