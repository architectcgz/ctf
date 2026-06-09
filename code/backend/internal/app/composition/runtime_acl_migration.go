package composition

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func migrateLegacyInstanceACLHandles(
	ctx context.Context,
	repo *runtimeinfra.Repository,
	router *runtimeNodeExecutionRouter,
	defaultClient *nodeRuntimeClient,
	logger *zap.Logger,
) error {
	if repo == nil {
		return nil
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	instances, err := repo.ListInstancesNeedingACLHandleMigration(ctx)
	if err != nil {
		return fmt.Errorf("list instances needing acl migration: %w", err)
	}
	for i := range instances {
		if err := migrateLegacyInstanceACLHandle(
			ctx,
			repo,
			router,
			defaultClient,
			instanceFromRuntimeManaged(&instances[i]),
		); err != nil {
			return err
		}
	}
	if len(instances) > 0 {
		logger.Info("legacy instance acl handles migrated", zap.Int("count", len(instances)))
	}
	return nil
}

func migrateLegacyInstanceACLHandle(
	ctx context.Context,
	repo *runtimeinfra.Repository,
	router *runtimeNodeExecutionRouter,
	defaultClient *nodeRuntimeClient,
	instance *instancecontracts.Instance,
) error {
	if repo == nil || instance == nil || instance.ID <= 0 {
		return nil
	}

	details, err := runtimecontracts.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err != nil {
		return fmt.Errorf("decode runtime details for instance %d: %w", instance.ID, err)
	}
	if details.ACL != nil || len(details.ACLRules) == 0 {
		return nil
	}

	handle := &runtimecontracts.InstanceRuntimeACLHandle{
		Chain: fmt.Sprintf("CTF-INS-%d", instance.ID),
	}

	client, err := runtimeNodeClientForInstanceMigration(ctx, router, defaultClient, instance)
	if err != nil {
		return fmt.Errorf("resolve runtime node client for instance %d: %w", instance.ID, err)
	}
	if err := client.ApplyACL(ctx, handle, details.ACLRules); err != nil {
		return fmt.Errorf("apply legacy acl handle for instance %d: %w", instance.ID, err)
	}
	if err := client.RemoveACLRules(ctx, details.ACLRules); err != nil && !shouldIgnoreLegacyACLRemovalError(err) {
		return fmt.Errorf("remove legacy acl rules for instance %d: %w", instance.ID, err)
	}

	details.ACL = handle
	encodedDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(details)
	if err != nil {
		return fmt.Errorf("encode migrated runtime details for instance %d: %w", instance.ID, err)
	}
	if err := repo.UpdateInstanceRuntimeDetails(ctx, instance.ID, encodedDetails); err != nil {
		return fmt.Errorf("persist migrated runtime details for instance %d: %w", instance.ID, err)
	}
	return nil
}

func runtimeNodeClientForInstanceMigration(
	ctx context.Context,
	router *runtimeNodeExecutionRouter,
	defaultClient *nodeRuntimeClient,
	instance *instancecontracts.Instance,
) (runtimeNodeClient, error) {
	if router != nil {
		client, _, err := router.clientForInstance(ctx, instance)
		if err != nil {
			return nil, err
		}
		if client != nil {
			return client, nil
		}
	}
	if defaultClient != nil {
		return defaultClient, nil
	}
	return nil, runtimeports.ErrRuntimeNodeUnavailable
}

func instanceFromRuntimeManaged(instance *runtimecontracts.RuntimeManagedInstance) *instancecontracts.Instance {
	if instance == nil {
		return nil
	}
	return &instancecontracts.Instance{
		ID:             instance.ID,
		UserID:         instance.UserID,
		ContestID:      instance.ContestID,
		TeamID:         instance.TeamID,
		ChallengeID:    instance.ChallengeID,
		ServiceID:      instance.ServiceID,
		NodeID:         instance.NodeID,
		HostPort:       instance.HostPort,
		ContainerID:    instance.ContainerID,
		NetworkID:      instance.NetworkID,
		RuntimeDetails: instance.RuntimeDetails,
		ShareScope:     instancecontracts.ShareScope(instance.ShareScope),
		Status:         instance.Status,
		AccessURL:      instance.AccessURL,
		Nonce:          instance.Nonce,
		FlagKeyID:      instance.FlagKeyID,
		ExpiresAt:      instance.ExpiresAt,
		DestroyedAt:    instance.DestroyedAt,
		ExtendCount:    instance.ExtendCount,
		MaxExtends:     instance.MaxExtends,
		CreatedAt:      instance.CreatedAt,
		UpdatedAt:      instance.UpdatedAt,
	}
}

func shouldIgnoreLegacyACLRemovalError(err error) bool {
	if err == nil {
		return false
	}
	lowerErr := strings.ToLower(err.Error())
	return strings.Contains(lowerErr, "does a matching rule exist") ||
		strings.Contains(lowerErr, "no chain/target/match by that name")
}
