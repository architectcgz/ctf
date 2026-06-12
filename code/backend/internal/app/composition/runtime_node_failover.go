package composition

import (
	"context"

	runtimeentity "ctf-platform/internal/module/container_runtime/entity"
)

type desiredAWDRuntimeReconciler interface {
	ReconcileDesiredAWDInstances(ctx context.Context) error
}

func WireRuntimeNodeFailover(runtime *ContainerRuntimeModule, instance *InstanceModule, practice *PracticeModule) {
	if runtime == nil {
		return
	}
	var reconciler desiredAWDRuntimeReconciler
	if practice != nil {
		reconciler = practice.AWDDesiredRuntimeReconciler
	}
	runtime.SetRuntimeNodeOfflineHandler(func(ctx context.Context, node runtimeentity.RuntimeNode) error {
		if instance != nil {
			if err := instance.HandleRuntimeNodeOffline(ctx, node.ID); err != nil {
				return err
			}
		}
		if reconciler != nil {
			return reconciler.ReconcileDesiredAWDInstances(ctx)
		}
		return nil
	})
}
