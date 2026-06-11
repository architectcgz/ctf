package challengepackageexport

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

func (s *ChallengePackageExportService) ExportChallengePackage(
	ctx context.Context,
	actorUserID int64,
	challengeID int64,
) (*challengecontracts.ChallengePackageExportResp, error) {
	var response *challengecontracts.ChallengePackageExportResp
	cleanupPaths := make([]string, 0, 2)
	if s.packageExportTxRunner == nil {
		return nil, fmt.Errorf("challenge package export tx runner is not configured")
	}
	if s.packageStorage == nil {
		return nil, fmt.Errorf("challenge package storage is not configured")
	}
	if err := s.packageExportTxRunner.WithinChallengePackageExportTransaction(ctx, func(store challengeports.ChallengePackageExportTxStore) error {
		challenge, err := store.FindChallenge(ctx, challengeID)
		if err != nil {
			if errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
				return challengecontracts.ErrChallengeNotFound
			}
			return err
		}

		topology, err := store.FindTopology(ctx, challengeID)
		if err != nil {
			if errors.Is(err, challengeports.ErrChallengeTopologyNotFound) {
				return apperror.ErrNotFound.WithCause(errors.New("题目拓扑不存在"))
			}
			return err
		}
		if topology.PackageRevisionID == nil || *topology.PackageRevisionID <= 0 {
			return apperror.ErrConflict.WithCause(errors.New("当前题目没有可导出的题包基线"))
		}

		baseRevision, err := store.FindPackageRevisionByID(ctx, *topology.PackageRevisionID)
		if err != nil {
			if errors.Is(err, challengeports.ErrChallengeTopologyPackageRevisionNotFound) {
				return apperror.ErrConflict.WithCause(errors.New("题包基线修订不存在"))
			}
			return err
		}
		if strings.TrimSpace(baseRevision.SourceDir) == "" {
			return apperror.ErrConflict.WithCause(errors.New("题包基线源码目录缺失"))
		}

		revisionNo, err := store.NextPackageRevisionNo(ctx, challengeID)
		if err != nil {
			return err
		}
		fileName := resolveChallengePackageSlug(challenge.PackageSlug, challenge.ID, baseRevision.PackageSlug) + ".zip"
		workspace, err := s.packageStorage.PrepareExportWorkspace(ctx, challengeports.ChallengePackageExportWorkspaceRequest{
			ChallengeID: challengeID,
			RevisionNo:  revisionNo,
			PackageSlug: resolveChallengePackageSlug(challenge.PackageSlug, challenge.ID, baseRevision.PackageSlug),
			SourceDir:   baseRevision.SourceDir,
			FileName:    fileName,
		})
		if err != nil {
			return err
		}
		cleanupPaths = append(cleanupPaths, workspace.ExportRoot)

		hints, err := store.ListChallengeHints(ctx, challengeID)
		if err != nil {
			return err
		}
		manifestRaw, err := rewriteChallengeManifestSnapshot(ctx, s.packageStorage, store, workspace.SourceDir, challenge, topology, hints, baseRevision)
		if err != nil {
			return err
		}
		topologyRaw, err := rewriteChallengeTopologySnapshot(ctx, s.packageStorage, store, workspace.SourceDir, topology, baseRevision)
		if err != nil {
			return err
		}

		if err := s.packageStorage.BuildExportArchive(ctx, *workspace); err != nil {
			return fmt.Errorf("zip exported package: %w", err)
		}

		now := time.Now().UTC()
		parentRevisionID := baseRevision.ID
		revision := &challengeentity.ChallengePackageRevision{
			ChallengeID:        challengeID,
			RevisionNo:         revisionNo,
			SourceType:         challengeentity.ChallengePackageRevisionSourceExported,
			ParentRevisionID:   &parentRevisionID,
			PackageSlug:        resolveChallengePackageSlug(challenge.PackageSlug, challenge.ID, baseRevision.PackageSlug),
			ArchivePath:        workspace.ArchivePath,
			SourceDir:          workspace.SourceDir,
			ManifestSnapshot:   manifestRaw,
			TopologySourcePath: resolveRevisionTopologySourcePath(topology, baseRevision),
			TopologySnapshot:   topologyRaw,
			CreatedBy:          int64Ptr(actorUserID),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := store.CreateExportRevision(ctx, revision); err != nil {
			return err
		}

		revisionID := revision.ID
		if err := store.MarkTopologyExported(ctx, topology.ID, revisionID, topology.Spec, now); err != nil {
			return err
		}

		response = &challengecontracts.ChallengePackageExportResp{
			ChallengeID: challengeID,
			RevisionID:  revision.ID,
			ArchivePath: workspace.ArchivePath,
			SourceDir:   workspace.SourceDir,
			FileName:    workspace.FileName,
			CreatedAt:   now,
		}
		return nil
	}); err != nil {
		for _, cleanupPath := range cleanupPaths {
			if strings.TrimSpace(cleanupPath) == "" {
				continue
			}
			_ = s.packageStorage.DeletePath(ctx, cleanupPath)
		}
		return nil, err
	}

	return response, nil
}

func (s *ChallengePackageExportService) GetChallengePackageExport(ctx context.Context, challengeID int64, revisionID *int64) (*challengecontracts.ChallengePackageExportResp, error) {
	if s.packageRepo == nil {
		return nil, apperror.ErrNotFound.WithCause(errors.New("题包修订仓储未配置"))
	}
	if s.packageStorage == nil {
		return nil, apperror.ErrNotFound.WithCause(errors.New("题包存储未配置"))
	}
	if _, err := s.challengeRepo.FindByID(ctx, challengeID); err != nil {
		if errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
			return nil, challengecontracts.ErrChallengeNotFound
		}
		return nil, err
	}

	var revision *challengeentity.ChallengePackageRevision
	var err error
	if revisionID != nil && *revisionID > 0 {
		revision, err = s.packageRepo.FindChallengePackageRevisionByID(ctx, *revisionID)
		if err != nil {
			if errors.Is(err, challengeports.ErrChallengeTopologyPackageRevisionNotFound) {
				return nil, apperror.ErrNotFound.WithCause(errors.New("题包修订不存在"))
			}
			return nil, err
		}
		if revision.ChallengeID != challengeID {
			return nil, apperror.ErrForbidden
		}
	} else {
		topology, findErr := s.topologyRepo.FindChallengeTopologyByChallengeID(ctx, challengeID)
		if findErr != nil {
			if errors.Is(findErr, challengeports.ErrChallengeTopologyNotFound) {
				return nil, apperror.ErrNotFound.WithCause(errors.New("题目拓扑不存在"))
			}
			return nil, findErr
		}
		selectedRevisionID := topology.LastExportRevisionID
		if selectedRevisionID == nil || *selectedRevisionID <= 0 {
			selectedRevisionID = topology.PackageRevisionID
		}
		if selectedRevisionID == nil || *selectedRevisionID <= 0 {
			return nil, apperror.ErrNotFound.WithCause(errors.New("尚未生成可下载的题包"))
		}
		revision, err = s.packageRepo.FindChallengePackageRevisionByID(ctx, *selectedRevisionID)
		if err != nil {
			if errors.Is(err, challengeports.ErrChallengeTopologyPackageRevisionNotFound) {
				return nil, apperror.ErrNotFound.WithCause(errors.New("题包修订不存在"))
			}
			return nil, err
		}
	}

	if strings.TrimSpace(revision.ArchivePath) == "" {
		return nil, apperror.ErrNotFound.WithCause(errors.New("当前修订没有可下载的题包归档"))
	}
	fileName, err := s.packageStorage.EnsureArchiveExists(ctx, revision.ArchivePath)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrNotFound.WithCause(errors.New("题包归档文件不存在"))
		}
		return nil, err
	}

	return &challengecontracts.ChallengePackageExportResp{
		ChallengeID: challengeID,
		RevisionID:  revision.ID,
		ArchivePath: revision.ArchivePath,
		SourceDir:   revision.SourceDir,
		FileName:    fileName,
		DownloadURL: "",
		CreatedAt:   revision.CreatedAt,
	}, nil
}

func rewriteChallengeManifestSnapshot(
	ctx context.Context,
	storage challengeports.ChallengePackageStorage,
	store challengeports.ChallengePackageExportTxStore,
	sourceDir string,
	challenge *challengeports.ChallengePackageCore,
	topology *challengeentity.ChallengeTopology,
	hints []challengeentity.ChallengeHint,
	revision *challengeentity.ChallengePackageRevision,
) (string, error) {
	var manifest domain.ChallengePackageManifest
	manifestRaw := strings.TrimSpace(revision.ManifestSnapshot)
	if manifestRaw == "" {
		content, err := storage.ReadTextFile(ctx, sourceDir, "challenge.yml")
		if err != nil {
			return "", err
		}
		manifestRaw = content
	}
	if err := yaml.Unmarshal([]byte(manifestRaw), &manifest); err != nil {
		return "", err
	}

	manifest.APIVersion = "v1"
	manifest.Kind = "challenge"
	manifest.Meta.Slug = resolveChallengePackageSlug(challenge.PackageSlug, challenge.ID, revision.PackageSlug)
	manifest.Meta.Title = challenge.Title
	manifest.Meta.Category = challenge.Category
	manifest.Meta.Difficulty = challenge.Difficulty
	manifest.Meta.Points = challenge.Points
	manifest.Flag.Type = challenge.FlagType
	manifest.Flag.Prefix = challenge.FlagPrefix
	switch challenge.FlagType {
	case challengecontracts.FlagTypeRegex:
		manifest.Flag.Value = challenge.FlagRegex
	case challengecontracts.FlagTypeDynamic, challengecontracts.FlagTypeManualReview:
		manifest.Flag.Value = ""
	}
	if challenge.ImageID > 0 {
		ref, err := store.FindImageRefByID(ctx, challenge.ImageID)
		if err != nil {
			return "", err
		}
		manifest.Runtime.Type = "container"
		manifest.Runtime.Image.Ref = ref
		if manifest.Runtime.Image.Name == "" {
			manifest.Runtime.Image.Name = ref
		}
	}
	if topology != nil && strings.TrimSpace(topology.SourcePath) != "" {
		manifest.Extensions.Topology.Enabled = true
		manifest.Extensions.Topology.Source = topology.SourcePath
	}
	if len(hints) > 0 {
		manifest.Hints = make([]domain.ChallengePackageHint, 0, len(hints))
		for _, hint := range hints {
			manifest.Hints = append(manifest.Hints, domain.ChallengePackageHint{
				Level:   hint.Level,
				Title:   hint.Title,
				Content: hint.Content,
			})
		}
	} else {
		manifest.Hints = nil
	}

	statementPath := strings.TrimSpace(manifest.Content.Statement)
	if statementPath == "" {
		statementPath = "statement.md"
		manifest.Content.Statement = statementPath
	}
	if err := storage.WriteTextFile(ctx, sourceDir, statementPath, challenge.Description); err != nil {
		return "", err
	}

	content, err := yaml.Marshal(&manifest)
	if err != nil {
		return "", err
	}
	if err := storage.WriteTextFile(ctx, sourceDir, "challenge.yml", string(content)); err != nil {
		return "", err
	}
	return string(content), nil
}

func rewriteChallengeTopologySnapshot(
	ctx context.Context,
	storage challengeports.ChallengePackageStorage,
	store challengeports.ChallengePackageExportTxStore,
	sourceDir string,
	topology *challengeentity.ChallengeTopology,
	revision *challengeentity.ChallengePackageRevision,
) (string, error) {
	if topology == nil {
		return "", nil
	}
	spec, err := challengecontracts.DecodeTopologySpec(topology.Spec)
	if err != nil {
		return "", err
	}

	var baseline domain.ChallengePackageTopologyManifest
	if raw := strings.TrimSpace(revision.TopologySnapshot); raw != "" {
		if err := yaml.Unmarshal([]byte(raw), &baseline); err != nil {
			return "", err
		}
	}
	baselineNodeImages := make(map[string]domain.ChallengePackageTopologyNodeImage, len(baseline.Nodes))
	for _, node := range baseline.Nodes {
		baselineNodeImages[node.Key] = node.Image
	}

	manifest := domain.ChallengePackageTopologyManifest{
		APIVersion:   "v1",
		Kind:         "topology",
		EntryNodeKey: topology.EntryNodeKey,
		Networks:     make([]domain.ChallengePackageTopologyNetwork, 0, len(spec.Networks)),
		Nodes:        make([]domain.ChallengePackageTopologyNode, 0, len(spec.Nodes)),
		Links:        make([]domain.ChallengePackageTopologyLink, 0, len(spec.Links)),
		Policies:     make([]domain.ChallengePackageTopologyPolicy, 0, len(spec.Policies)),
	}
	for _, network := range spec.Networks {
		manifest.Networks = append(manifest.Networks, domain.ChallengePackageTopologyNetwork{
			Key:      network.Key,
			Name:     network.Name,
			CIDR:     network.CIDR,
			Internal: network.Internal,
		})
	}
	for _, node := range spec.Nodes {
		image := baselineNodeImages[node.Key]
		if node.ImageID > 0 {
			ref, err := store.FindImageRefByID(ctx, node.ImageID)
			if err != nil {
				return "", err
			}
			image.Ref = ref
		}
		if strings.TrimSpace(image.Ref) == "" {
			return "", apperror.ErrInvalidParams.WithCause(fmt.Errorf("节点 %s 缺少镜像引用，无法导出题包", node.Key))
		}
		var resources *domain.ChallengePackageTopologyResources
		if node.Resources != nil {
			resources = &domain.ChallengePackageTopologyResources{
				CPUQuota:  node.Resources.CPUQuota,
				MemoryMB:  node.Resources.MemoryMB,
				PidsLimit: node.Resources.PidsLimit,
			}
		}
		manifest.Nodes = append(manifest.Nodes, domain.ChallengePackageTopologyNode{
			Key:         node.Key,
			Name:        node.Name,
			Tier:        node.Tier,
			Image:       image,
			ServicePort: node.ServicePort,
			InjectFlag:  node.InjectFlag,
			NetworkKeys: append([]string(nil), node.NetworkKeys...),
			Env:         node.Env,
			Resources:   resources,
		})
	}
	for _, link := range spec.Links {
		manifest.Links = append(manifest.Links, domain.ChallengePackageTopologyLink{
			FromNodeKey: link.FromNodeKey,
			ToNodeKey:   link.ToNodeKey,
		})
	}
	for _, policy := range spec.Policies {
		manifest.Policies = append(manifest.Policies, domain.ChallengePackageTopologyPolicy{
			SourceNodeKey: policy.SourceNodeKey,
			TargetNodeKey: policy.TargetNodeKey,
			Action:        policy.Action,
			Protocol:      policy.Protocol,
			Ports:         append([]int(nil), policy.Ports...),
		})
	}

	content, err := yaml.Marshal(&manifest)
	if err != nil {
		return "", err
	}
	topologyPath := resolveRevisionTopologySourcePath(topology, revision)
	if err := storage.WriteTextFile(ctx, sourceDir, topologyPath, string(content)); err != nil {
		return "", err
	}
	return string(content), nil
}

func resolveChallengePackageSlug(packageSlug *string, challengeID int64, fallback string) string {
	if packageSlug != nil && strings.TrimSpace(*packageSlug) != "" {
		return strings.TrimSpace(*packageSlug)
	}
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}
	if challengeID > 0 {
		return fmt.Sprintf("challenge-%d", challengeID)
	}
	return "challenge-package"
}

func resolveTopologySourcePath(topology *domain.ParsedChallengePackageTopology) string {
	if topology == nil {
		return ""
	}
	return strings.TrimSpace(topology.Source)
}

func resolveTopologySnapshot(topology *domain.ParsedChallengePackageTopology) string {
	if topology == nil {
		return ""
	}
	return topology.Raw
}

func resolveRevisionTopologySourcePath(topology *challengeentity.ChallengeTopology, revision *challengeentity.ChallengePackageRevision) string {
	if topology != nil && strings.TrimSpace(topology.SourcePath) != "" {
		return strings.TrimSpace(topology.SourcePath)
	}
	if revision != nil && strings.TrimSpace(revision.TopologySourcePath) != "" {
		return strings.TrimSpace(revision.TopologySourcePath)
	}
	return "docker/topology.yml"
}

func int64Ptr(value int64) *int64 {
	return &value
}
