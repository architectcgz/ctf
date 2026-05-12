package composition

import (
	challengeruntime "ctf-platform/internal/module/challenge/runtime"
)

type BackgroundTaskCloser = challengeruntime.BackgroundTaskCloser
type ChallengeModule = challengeruntime.Module

func BuildChallengeModule(root *Root, runtime *ContainerRuntimeModule) (*ChallengeModule, error) {
	module, err := challengeruntime.Build(challengeruntime.Deps{
		AppContext:   root.Context(),
		Config:       root.Config(),
		Logger:       root.Logger(),
		DB:           root.DB(),
		Cache:        root.Cache(),
		Events:       root.Events,
		ImageRuntime: runtime.ChallengeImageRuntime,
		RuntimeProbe: runtime.ChallengeRuntimeProbe,
	})
	if err != nil {
		return nil, err
	}

	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewLoopBackgroundJob(job.Name, job.Run))
	}
	return module, nil
}
