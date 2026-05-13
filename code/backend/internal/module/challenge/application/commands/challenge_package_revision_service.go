package commands

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

func (s *ChallengeService) createImportedPackageRevision(
	tx *gorm.DB,
	actorUserID int64,
	challenge *model.Challenge,
	record storedChallengeImportPreview,
	parsed *domain.ParsedChallengePackage,
) (*model.ChallengePackageRevision, error) {
	if challenge == nil || parsed == nil {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("缺少题目或题包信息"))
	}

	revisionNo, err := nextChallengePackageRevisionNo(tx, challenge.ID)
	if err != nil {
		return nil, err
	}
	revisionRoot := filepath.Join(challengePackageSourceRoot(), fmt.Sprintf("challenge-%d", challenge.ID), fmt.Sprintf("r%04d", revisionNo))
	sourceDir := filepath.Join(revisionRoot, "source")
	if err := copyDirectoryTree(parsed.RootDir, sourceDir); err != nil {
		return nil, fmt.Errorf("copy imported package source: %w", err)
	}

	archivePath := ""
	previewArchivePath := filepath.Join(challengeImportPreviewRoot(), record.ID, "package.zip")
	if info, statErr := os.Stat(previewArchivePath); statErr == nil && !info.IsDir() {
		archivePath = filepath.Join(revisionRoot, sanitizeImportedAttachmentName(record.FileName, "challenge-package.zip"))
		if err := copyFile(previewArchivePath, archivePath); err != nil {
			return nil, fmt.Errorf("copy imported package archive: %w", err)
		}
	}

	now := time.Now()
	revision := &model.ChallengePackageRevision{
		ChallengeID:        challenge.ID,
		RevisionNo:         revisionNo,
		SourceType:         model.ChallengePackageRevisionSourceImported,
		PackageSlug:        resolveChallengePackageSlug(challenge, parsed.Slug),
		ArchivePath:        archivePath,
		SourceDir:          sourceDir,
		ManifestSnapshot:   parsed.ManifestRaw,
		TopologySourcePath: resolveTopologySourcePath(parsed.Topology),
		TopologySnapshot:   resolveTopologySnapshot(parsed.Topology),
		CreatedBy:          int64Ptr(actorUserID),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := tx.Create(revision).Error; err != nil {
		return nil, err
	}
	return revision, nil
}

func (s *ChallengeService) ExportChallengePackage(
	ctx context.Context,
	actorUserID int64,
	challengeID int64,
) (*dto.ChallengePackageExportResp, error) {
	var response *dto.ChallengePackageExportResp
	cleanupPaths := make([]string, 0, 2)
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		challenge, err := findChallengeByIDTx(tx, challengeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrChallengeNotFound
			}
			return err
		}

		topology, err := findChallengeTopologyByChallengeIDTx(tx, challengeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrNotFound.WithCause(errors.New("题目拓扑不存在"))
			}
			return err
		}
		if topology.PackageRevisionID == nil || *topology.PackageRevisionID <= 0 {
			return errcode.ErrConflict.WithCause(errors.New("当前题目没有可导出的题包基线"))
		}

		baseRevision, err := findChallengePackageRevisionByIDTx(tx, *topology.PackageRevisionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrConflict.WithCause(errors.New("题包基线修订不存在"))
			}
			return err
		}
		if strings.TrimSpace(baseRevision.SourceDir) == "" {
			return errcode.ErrConflict.WithCause(errors.New("题包基线源码目录缺失"))
		}

		revisionNo, err := nextChallengePackageRevisionNo(tx, challengeID)
		if err != nil {
			return err
		}
		exportRoot := filepath.Join(challengePackageExportRoot(), fmt.Sprintf("challenge-%d", challengeID), fmt.Sprintf("r%04d", revisionNo))
		sourceDir := filepath.Join(exportRoot, "source")
		if err := copyDirectoryTree(baseRevision.SourceDir, sourceDir); err != nil {
			return fmt.Errorf("copy export package source: %w", err)
		}
		cleanupPaths = append(cleanupPaths, sourceDir)

		hints, err := listChallengeHints(tx, challengeID)
		if err != nil {
			return err
		}
		manifestRaw, err := rewriteChallengeManifestSnapshot(tx, sourceDir, challenge, topology, hints, baseRevision)
		if err != nil {
			return err
		}
		topologyRaw, err := rewriteChallengeTopologySnapshot(tx, sourceDir, topology, baseRevision)
		if err != nil {
			return err
		}

		fileName := sanitizeImportedAttachmentName(resolveChallengePackageSlug(challenge, baseRevision.PackageSlug)+".zip", "challenge-package.zip")
		archivePath := filepath.Join(exportRoot, fileName)
		if err := zipDirectory(sourceDir, archivePath); err != nil {
			return fmt.Errorf("zip exported package: %w", err)
		}
		cleanupPaths = append(cleanupPaths, archivePath)

		now := time.Now()
		parentRevisionID := baseRevision.ID
		revision := &model.ChallengePackageRevision{
			ChallengeID:        challengeID,
			RevisionNo:         revisionNo,
			SourceType:         model.ChallengePackageRevisionSourceExported,
			ParentRevisionID:   &parentRevisionID,
			PackageSlug:        resolveChallengePackageSlug(challenge, baseRevision.PackageSlug),
			ArchivePath:        archivePath,
			SourceDir:          sourceDir,
			ManifestSnapshot:   manifestRaw,
			TopologySourcePath: resolveRevisionTopologySourcePath(topology, baseRevision),
			TopologySnapshot:   topologyRaw,
			CreatedBy:          int64Ptr(actorUserID),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := tx.Create(revision).Error; err != nil {
			return err
		}

		revisionID := revision.ID
		if err := tx.Model(&model.ChallengeTopology{}).
			Where("id = ?", topology.ID).
			Updates(map[string]any{
				"package_revision_id":     revisionID,
				"package_baseline_spec":   topology.Spec,
				"sync_status":             model.ChallengeTopologySyncStatusClean,
				"last_export_revision_id": revisionID,
				"updated_at":              now,
			}).Error; err != nil {
			return err
		}

		response = &dto.ChallengePackageExportResp{
			ChallengeID: challengeID,
			RevisionID:  revision.ID,
			ArchivePath: archivePath,
			SourceDir:   sourceDir,
			FileName:    fileName,
			CreatedAt:   now,
		}
		return nil
	}); err != nil {
		for _, cleanupPath := range cleanupPaths {
			if strings.TrimSpace(cleanupPath) == "" {
				continue
			}
			_ = os.RemoveAll(cleanupPath)
		}
		return nil, err
	}

	return response, nil
}

func (s *ChallengeService) GetChallengePackageExport(ctx context.Context, challengeID int64, revisionID *int64) (*dto.ChallengePackageExportResp, error) {
	if s.packageRepo == nil {
		return nil, errcode.ErrNotFound.WithCause(errors.New("题包修订仓储未配置"))
	}
	if _, err := s.repo.FindByID(ctx, challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, challengeports.ErrChallengeCommandChallengeNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}

	var revision *model.ChallengePackageRevision
	var err error
	if revisionID != nil && *revisionID > 0 {
		revision, err = s.packageRepo.FindChallengePackageRevisionByID(ctx, *revisionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrNotFound.WithCause(errors.New("题包修订不存在"))
			}
			return nil, err
		}
		if revision.ChallengeID != challengeID {
			return nil, errcode.ErrForbidden
		}
	} else {
		topology, findErr := s.topologyRepo.FindChallengeTopologyByChallengeID(ctx, challengeID)
		if findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) || errors.Is(findErr, challengeports.ErrChallengeTopologyNotFound) {
				return nil, errcode.ErrNotFound.WithCause(errors.New("题目拓扑不存在"))
			}
			return nil, findErr
		}
		selectedRevisionID := topology.LastExportRevisionID
		if selectedRevisionID == nil || *selectedRevisionID <= 0 {
			selectedRevisionID = topology.PackageRevisionID
		}
		if selectedRevisionID == nil || *selectedRevisionID <= 0 {
			return nil, errcode.ErrNotFound.WithCause(errors.New("尚未生成可下载的题包"))
		}
		revision, err = s.packageRepo.FindChallengePackageRevisionByID(ctx, *selectedRevisionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrNotFound.WithCause(errors.New("题包修订不存在"))
			}
			return nil, err
		}
	}

	if strings.TrimSpace(revision.ArchivePath) == "" {
		return nil, errcode.ErrNotFound.WithCause(errors.New("当前修订没有可下载的题包归档"))
	}
	if _, err := os.Stat(revision.ArchivePath); err != nil {
		if os.IsNotExist(err) {
			return nil, errcode.ErrNotFound.WithCause(errors.New("题包归档文件不存在"))
		}
		return nil, err
	}

	resp := challengeCommandResponseMapperInst.ToChallengePackageExportRespBasePtr(revision)
	resp.ChallengeID = challengeID
	resp.FileName = filepath.Base(revision.ArchivePath)
	return resp, nil
}

func nextChallengePackageRevisionNo(tx *gorm.DB, challengeID int64) (int, error) {
	var latest model.ChallengePackageRevision
	err := tx.Where("challenge_id = ?", challengeID).Order("revision_no DESC, id DESC").First(&latest).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return 1, nil
	case err != nil:
		return 0, err
	default:
		return latest.RevisionNo + 1, nil
	}
}

func findChallengeByIDTx(tx *gorm.DB, challengeID int64) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := tx.Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func findChallengeTopologyByChallengeIDTx(tx *gorm.DB, challengeID int64) (*model.ChallengeTopology, error) {
	var topology model.ChallengeTopology
	if err := tx.Where("challenge_id = ?", challengeID).First(&topology).Error; err != nil {
		return nil, err
	}
	return &topology, nil
}

func findChallengePackageRevisionByIDTx(tx *gorm.DB, revisionID int64) (*model.ChallengePackageRevision, error) {
	var revision model.ChallengePackageRevision
	if err := tx.Where("id = ?", revisionID).First(&revision).Error; err != nil {
		return nil, err
	}
	return &revision, nil
}

func upsertChallengeTopologyTx(tx *gorm.DB, topology *model.ChallengeTopology) error {
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "challenge_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"template_id",
			"entry_node_key",
			"spec",
			"source_type",
			"source_path",
			"package_revision_id",
			"package_baseline_spec",
			"sync_status",
			"last_export_revision_id",
			"updated_at",
			"deleted_at",
		}),
	}).Create(topology).Error
}

func rewriteChallengeManifestSnapshot(
	tx *gorm.DB,
	sourceDir string,
	challenge *model.Challenge,
	topology *model.ChallengeTopology,
	hints []model.ChallengeHint,
	revision *model.ChallengePackageRevision,
) (string, error) {
	var manifest domain.ChallengePackageManifest
	manifestRaw := strings.TrimSpace(revision.ManifestSnapshot)
	if manifestRaw == "" {
		content, err := os.ReadFile(filepath.Join(sourceDir, "challenge.yml"))
		if err != nil {
			return "", err
		}
		manifestRaw = string(content)
	}
	if err := yaml.Unmarshal([]byte(manifestRaw), &manifest); err != nil {
		return "", err
	}

	manifest.APIVersion = "v1"
	manifest.Kind = "challenge"
	manifest.Meta.Slug = resolveChallengePackageSlug(challenge, revision.PackageSlug)
	manifest.Meta.Title = challenge.Title
	manifest.Meta.Category = challenge.Category
	manifest.Meta.Difficulty = challenge.Difficulty
	manifest.Meta.Points = challenge.Points
	manifest.Flag.Type = challenge.FlagType
	manifest.Flag.Prefix = challenge.FlagPrefix
	switch challenge.FlagType {
	case model.FlagTypeRegex:
		manifest.Flag.Value = challenge.FlagRegex
	case model.FlagTypeDynamic, model.FlagTypeManualReview:
		manifest.Flag.Value = ""
	}
	if challenge.ImageID > 0 {
		ref, err := findImageRefByID(tx, challenge.ImageID)
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
	absoluteStatementPath := filepath.Join(sourceDir, filepath.FromSlash(statementPath))
	if err := os.MkdirAll(filepath.Dir(absoluteStatementPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(absoluteStatementPath, []byte(challenge.Description), 0o644); err != nil {
		return "", err
	}

	content, err := yaml.Marshal(&manifest)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "challenge.yml"), content, 0o644); err != nil {
		return "", err
	}
	return string(content), nil
}

func rewriteChallengeTopologySnapshot(
	tx *gorm.DB,
	sourceDir string,
	topology *model.ChallengeTopology,
	revision *model.ChallengePackageRevision,
) (string, error) {
	if topology == nil {
		return "", nil
	}
	spec, err := model.DecodeTopologySpec(topology.Spec)
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
			ref, err := findImageRefByID(tx, node.ImageID)
			if err != nil {
				return "", err
			}
			image.Ref = ref
		}
		if strings.TrimSpace(image.Ref) == "" {
			return "", errcode.ErrInvalidParams.WithCause(fmt.Errorf("节点 %s 缺少镜像引用，无法导出题包", node.Key))
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
	absolutePath := filepath.Join(sourceDir, filepath.FromSlash(topologyPath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(absolutePath, content, 0o644); err != nil {
		return "", err
	}
	return string(content), nil
}

func resolveChallengePackageSlug(challenge *model.Challenge, fallback string) string {
	if challenge != nil && challenge.PackageSlug != nil && strings.TrimSpace(*challenge.PackageSlug) != "" {
		return strings.TrimSpace(*challenge.PackageSlug)
	}
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}
	if challenge != nil && challenge.ID > 0 {
		return fmt.Sprintf("challenge-%d", challenge.ID)
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

func resolveRevisionTopologySourcePath(topology *model.ChallengeTopology, revision *model.ChallengePackageRevision) string {
	if topology != nil && strings.TrimSpace(topology.SourcePath) != "" {
		return strings.TrimSpace(topology.SourcePath)
	}
	if revision != nil && strings.TrimSpace(revision.TopologySourcePath) != "" {
		return strings.TrimSpace(revision.TopologySourcePath)
	}
	return "docker/topology.yml"
}

func listChallengeHints(tx *gorm.DB, challengeID int64) ([]model.ChallengeHint, error) {
	var hints []model.ChallengeHint
	if err := tx.Where("challenge_id = ?", challengeID).Order("level ASC, id ASC").Find(&hints).Error; err != nil {
		return nil, err
	}
	return hints, nil
}

func findImageRefByID(tx *gorm.DB, imageID int64) (string, error) {
	var image model.Image
	if err := tx.Where("id = ?", imageID).First(&image).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errcode.ErrInvalidParams.WithCause(errors.New("拓扑节点引用的镜像不存在"))
		}
		return "", err
	}
	if strings.TrimSpace(image.Name) == "" {
		return "", errcode.ErrInvalidParams.WithCause(errors.New("镜像记录缺少名称"))
	}
	if strings.TrimSpace(image.Tag) == "" || strings.TrimSpace(image.Tag) == "latest" {
		return strings.TrimSpace(image.Name), nil
	}
	return fmt.Sprintf("%s:%s", strings.TrimSpace(image.Name), strings.TrimSpace(image.Tag)), nil
}

func copyDirectoryTree(sourceDir string, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	return filepath.Walk(sourceDir, func(current string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceDir, current)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return nil
		}
		targetPath := filepath.Join(targetDir, relativePath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode().Perm())
		}
		return copyFile(current, targetPath)
	})
}

func copyFile(sourcePath string, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	info, err := source.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return err
	}
	return nil
}

func zipDirectory(sourceDir string, archivePath string) error {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return err
	}
	target, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer target.Close()

	writer := zip.NewWriter(target)
	defer writer.Close()

	entries := make([]string, 0, 32)
	if err := filepath.Walk(sourceDir, func(current string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(sourceDir, current)
		if err != nil {
			return err
		}
		entries = append(entries, relativePath)
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(entries)

	for _, relativePath := range entries {
		sourcePath := filepath.Join(sourceDir, relativePath)
		if err := addZipFile(writer, sourceDir, sourcePath); err != nil {
			return err
		}
	}
	return nil
}

func addZipFile(writer *zip.Writer, rootDir string, sourcePath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate
	relativePath, err := filepath.Rel(rootDir, sourcePath)
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(relativePath)
	fileWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(fileWriter, file)
	return err
}
