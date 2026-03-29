package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/infrastructure/postgres"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
)

const (
	defaultGlobalFlagSecret = "dev-integration-secret-123456789"
	defaultPacksDir         = "../../docs/challenges/packs"
)

type packManifest struct {
	Slug        string `yaml:"slug"`
	Title       string `yaml:"title"`
	Category    string `yaml:"category"`
	Difficulty  string `yaml:"difficulty"`
	Points      int    `yaml:"points"`
	Description struct {
		File string `yaml:"file"`
	} `yaml:"description"`
	Hints []packHint `yaml:"hints"`
	Flag  struct {
		Mode string `yaml:"mode"`
	} `yaml:"flag"`
	Runtime struct {
		Type  string `yaml:"type"`
		Image struct {
			Ref  string `yaml:"ref"`
			Name string `yaml:"name"`
			Tag  string `yaml:"tag"`
		} `yaml:"image"`
	} `yaml:"runtime"`
	Attachments []packAttachment `yaml:"attachments"`
}

type packHint struct {
	Text  string `yaml:"text"`
	Cost  int    `yaml:"cost"`
	Title string `yaml:"title"`
}

type packAttachment struct {
	Path string `yaml:"path"`
	File string `yaml:"file"`
	Name string `yaml:"name"`
}

type challengeSpec struct {
	Slug        string
	Title       string
	Description string
	Category    string
	Difficulty  string
	Points      int
	ImageID     int64
	ImageRef    string
	Attachment  string
	Hints       []hintSpec
	FlagPrefix  string
	FlagValue   string
}

type hintSpec struct {
	Level      int
	Title      string
	CostPoints int
	Content    string
}

type importResult struct {
	Total     int
	Created   int
	Updated   int
	Published int
	Failed    int
}

func main() {
	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if env == "" {
		env = "dev"
	}
	if strings.TrimSpace(os.Getenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET")) == "" {
		if err := os.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", defaultGlobalFlagSecret); err != nil {
			panic(fmt.Errorf("set default CTF_CONTAINER_FLAG_GLOBAL_SECRET: %w", err))
		}
	}

	cfg, err := config.Load(env)
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	db, err := postgres.Open(cfg.Postgres)
	if err != nil {
		panic(fmt.Errorf("open postgres: %w", err))
	}
	if getenvBool("CHALLENGE_IMPORT_AUTO_MIGRATE", false) {
		if err := db.AutoMigrate(&model.Image{}, &model.Challenge{}, &model.ChallengeHint{}); err != nil {
			panic(fmt.Errorf("auto migrate import schema: %w", err))
		}
	}

	packsDir := strings.TrimSpace(os.Getenv("CHALLENGE_PACKS_DIR"))
	if packsDir == "" {
		packsDir = defaultPacksDir
	}
	publish := getenvBool("CHALLENGE_IMPORT_PUBLISH", true)
	forceFlag := getenvBool("CHALLENGE_IMPORT_FORCE_FLAG", false)

	result, err := importChallengePacks(db, packsDir, publish, forceFlag)
	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"import completed: total=%d created=%d updated=%d published=%d failed=%d\n",
		result.Total,
		result.Created,
		result.Updated,
		result.Published,
		result.Failed,
	)
}

func importChallengePacks(db *gorm.DB, packsDir string, publish, forceFlag bool) (*importResult, error) {
	entries, err := os.ReadDir(packsDir)
	if err != nil {
		return nil, fmt.Errorf("read packs dir %s: %w", packsDir, err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	result := &importResult{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		packDir := filepath.Join(packsDir, entry.Name())
		manifestPath := filepath.Join(packDir, "manifest.yml")
		if _, err := os.Stat(manifestPath); err != nil {
			continue
		}

		result.Total++
		created, publishedNow, err := importOnePack(db, packDir, publish, forceFlag)
		if err != nil {
			result.Failed++
			fmt.Fprintf(os.Stderr, "[failed] %s: %v\n", entry.Name(), err)
			continue
		}
		if created {
			result.Created++
		} else {
			result.Updated++
		}
		if publishedNow {
			result.Published++
		}
	}

	return result, nil
}

func importOnePack(db *gorm.DB, packDir string, publish, forceFlag bool) (bool, bool, error) {
	manifest, err := readManifest(filepath.Join(packDir, "manifest.yml"))
	if err != nil {
		return false, false, err
	}

	spec, err := buildChallengeSpec(packDir, manifest)
	if err != nil {
		return false, false, err
	}

	created := false
	publishedNow := false
	if err := db.Transaction(func(tx *gorm.DB) error {
		resolvedImageID, err := resolveImportedImageID(tx, spec.Slug, spec.ImageRef)
		if err != nil {
			return err
		}

		var challenge model.Challenge
		err = tx.Unscoped().
			Where("title = ? AND category = ?", spec.Title, spec.Category).
			First(&challenge).Error

		now := time.Now()
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			challenge = model.Challenge{
				Title:         spec.Title,
				Description:   spec.Description,
				Category:      spec.Category,
				Difficulty:    spec.Difficulty,
				Points:        spec.Points,
				ImageID:       chooseImportedImageID(spec.ImageID, resolvedImageID),
				AttachmentURL: spec.Attachment,
				Status:        model.ChallengeStatusDraft,
				FlagPrefix:    spec.FlagPrefix,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := tx.Create(&challenge).Error; err != nil {
				return fmt.Errorf("create challenge %s: %w", spec.Slug, err)
			}
			created = true
		case err != nil:
			return fmt.Errorf("find challenge %s: %w", spec.Slug, err)
		default:
			imageID := challenge.ImageID
			if resolvedImageID > 0 {
				imageID = resolvedImageID
			} else if imageID == 0 && spec.ImageID > 0 {
				imageID = spec.ImageID
			}
			attachment := challenge.AttachmentURL
			if spec.Attachment != "" {
				attachment = spec.Attachment
			}

			updates := map[string]any{
				"description":    spec.Description,
				"difficulty":     spec.Difficulty,
				"points":         spec.Points,
				"image_id":       imageID,
				"attachment_url": attachment,
				"deleted_at":     nil,
				"updated_at":     now,
			}
			if err := tx.Model(&challenge).Updates(updates).Error; err != nil {
				return fmt.Errorf("update challenge %s: %w", spec.Slug, err)
			}
		}

		if err := syncChallengeHints(tx, challenge.ID, spec.Hints); err != nil {
			return err
		}

		needFlagConfig := created || forceFlag || strings.TrimSpace(challenge.FlagType) == "" ||
			(challenge.FlagType == model.FlagTypeStatic && strings.TrimSpace(challenge.FlagHash) == "")
		if needFlagConfig {
			if err := configureStaticFlag(tx, challenge.ID, spec.FlagPrefix, spec.FlagValue); err != nil {
				return err
			}
		}

		if publish && challenge.Status != model.ChallengeStatusPublished {
			if err := tx.Model(&model.Challenge{}).
				Where("id = ?", challenge.ID).
				Updates(map[string]any{
					"status":     model.ChallengeStatusPublished,
					"updated_at": time.Now(),
				}).Error; err != nil {
				return fmt.Errorf("publish challenge %s: %w", spec.Slug, err)
			}
			publishedNow = true
		}

		return nil
	}); err != nil {
		return false, false, err
	}

	return created, publishedNow, nil
}

func readManifest(path string) (*packManifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", path, err)
	}
	var manifest packManifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", path, err)
	}
	return &manifest, nil
}

func buildChallengeSpec(packDir string, manifest *packManifest) (*challengeSpec, error) {
	title := strings.TrimSpace(manifest.Title)
	if title == "" {
		return nil, fmt.Errorf("manifest %s missing title", packDir)
	}
	if strings.TrimSpace(manifest.Slug) == "" {
		return nil, fmt.Errorf("manifest %s missing slug", packDir)
	}

	descriptionFile := strings.TrimSpace(manifest.Description.File)
	if descriptionFile == "" {
		descriptionFile = "statement.md"
	}
	descriptionPath, err := safeJoin(packDir, descriptionFile)
	if err != nil {
		return nil, fmt.Errorf("resolve statement path for %s: %w", manifest.Slug, err)
	}
	descBytes, err := os.ReadFile(descriptionPath)
	if err != nil {
		return nil, fmt.Errorf("read statement for %s: %w", manifest.Slug, err)
	}
	description := strings.TrimSpace(string(descBytes))
	if description == "" {
		description = title
	}

	difficulty := normalizeDifficulty(manifest.Difficulty)
	points := manifest.Points
	if points <= 0 {
		points = defaultPointsByDifficulty(difficulty)
	}

	hints := make([]hintSpec, 0, len(manifest.Hints))
	for i, hint := range manifest.Hints {
		content := strings.TrimSpace(hint.Text)
		if content == "" {
			continue
		}
		title := strings.TrimSpace(hint.Title)
		if title == "" {
			title = fmt.Sprintf("Hint %d", i+1)
		}
		cost := hint.Cost
		if cost < 0 {
			cost = 0
		}
		hints = append(hints, hintSpec{
			Level:      len(hints) + 1,
			Title:      title,
			CostPoints: cost,
			Content:    content,
		})
	}

	flagValue := fmt.Sprintf("flag{%s}", sanitizeFlagToken(manifest.Slug))
	attachmentRel := firstAttachmentRelativePath(manifest.Attachments)
	attachmentURL := ""
	if attachmentRel != "" {
		attachmentPath, err := safeJoin(packDir, attachmentRel)
		if err != nil {
			return nil, fmt.Errorf("resolve attachment path for %s: %w", manifest.Slug, err)
		}
		if _, err := os.Stat(attachmentPath); err != nil {
			return nil, fmt.Errorf("attachment not found for %s: %w", manifest.Slug, err)
		}
		attachmentURL = buildAttachmentURL(manifest.Slug, attachmentRel)
	}

	return &challengeSpec{
		Slug:        manifest.Slug,
		Title:       title,
		Description: description,
		Category:    normalizeCategory(manifest.Category),
		Difficulty:  difficulty,
		Points:      points,
		ImageID:     0,
		ImageRef:    resolveRuntimeImageRef(manifest),
		Attachment:  attachmentURL,
		Hints:       hints,
		FlagPrefix:  "flag",
		FlagValue:   flagValue,
	}, nil
}

func resolveRuntimeImageRef(manifest *packManifest) string {
	if manifest == nil || strings.TrimSpace(manifest.Runtime.Type) != "container" {
		return ""
	}

	if ref := strings.TrimSpace(manifest.Runtime.Image.Ref); ref != "" {
		return ref
	}

	name := strings.TrimSpace(manifest.Runtime.Image.Name)
	if name == "" {
		return ""
	}
	tag := strings.TrimSpace(manifest.Runtime.Image.Tag)
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s:%s", name, tag)
}

func firstAttachmentRelativePath(attachments []packAttachment) string {
	for _, attachment := range attachments {
		rel := strings.TrimSpace(attachment.Path)
		if rel == "" {
			rel = strings.TrimSpace(attachment.File)
		}
		if rel != "" {
			return rel
		}
	}
	return ""
}

func buildAttachmentURL(slug, rel string) string {
	cleanRel := path.Clean("/" + strings.ReplaceAll(rel, "\\", "/"))
	cleanRel = strings.TrimPrefix(cleanRel, "/")

	segments := []string{"/api/v1/challenges/attachments", url.PathEscape(slug)}
	for _, part := range strings.Split(cleanRel, "/") {
		part = strings.TrimSpace(part)
		if part == "" || part == "." || part == ".." {
			continue
		}
		segments = append(segments, url.PathEscape(part))
	}
	return strings.Join(segments, "/")
}

func syncChallengeHints(tx *gorm.DB, challengeID int64, hints []hintSpec) error {
	if err := tx.Where("challenge_id = ?", challengeID).Delete(&model.ChallengeHint{}).Error; err != nil {
		return fmt.Errorf("delete hints for challenge %d: %w", challengeID, err)
	}
	if len(hints) == 0 {
		return nil
	}

	now := time.Now()
	records := make([]model.ChallengeHint, 0, len(hints))
	for _, hint := range hints {
		records = append(records, model.ChallengeHint{
			ChallengeID: challengeID,
			Level:       hint.Level,
			Title:       hint.Title,
			CostPoints:  hint.CostPoints,
			Content:     hint.Content,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	if err := tx.Create(&records).Error; err != nil {
		return fmt.Errorf("create hints for challenge %d: %w", challengeID, err)
	}
	return nil
}

func configureStaticFlag(tx *gorm.DB, challengeID int64, prefix, value string) error {
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("generate salt for challenge %d: %w", challengeID, err)
	}
	updates := map[string]any{
		"flag_type":   model.FlagTypeStatic,
		"flag_salt":   salt,
		"flag_hash":   crypto.HashStaticFlag(value, salt),
		"flag_prefix": prefix,
		"updated_at":  time.Now(),
	}
	if err := tx.Model(&model.Challenge{}).
		Where("id = ?", challengeID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("configure static flag for challenge %d: %w", challengeID, err)
	}
	return nil
}

func chooseImportedImageID(specImageID, resolvedImageID int64) int64 {
	if resolvedImageID > 0 {
		return resolvedImageID
	}
	return specImageID
}

func resolveImportedImageID(tx *gorm.DB, slug, imageRef string) (int64, error) {
	ref := strings.TrimSpace(imageRef)
	if ref == "" {
		return 0, nil
	}
	name, tag, err := splitImageRef(ref)
	if err != nil {
		return 0, fmt.Errorf("invalid image ref for %s: %w", slug, err)
	}

	var image model.Image
	findErr := tx.Unscoped().
		Where("name = ? AND tag = ?", name, tag).
		First(&image).Error
	switch {
	case errors.Is(findErr, gorm.ErrRecordNotFound):
		image = model.Image{
			Name:        name,
			Tag:         tag,
			Description: fmt.Sprintf("Imported from challenge pack %s", slug),
			Status:      model.ImageStatusAvailable,
			Size:        0,
		}
		if err := tx.Create(&image).Error; err != nil {
			return 0, fmt.Errorf("create image %s:%s for %s: %w", name, tag, slug, err)
		}
		return image.ID, nil
	case findErr != nil:
		return 0, fmt.Errorf("find image %s:%s for %s: %w", name, tag, slug, findErr)
	default:
		if err := tx.Model(&image).Updates(map[string]any{
			"status":     model.ImageStatusAvailable,
			"deleted_at": nil,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return 0, fmt.Errorf("update image %s:%s for %s: %w", name, tag, slug, err)
		}
		return image.ID, nil
	}
}

func splitImageRef(imageRef string) (string, string, error) {
	trimmed := strings.TrimSpace(imageRef)
	if trimmed == "" {
		return "", "", fmt.Errorf("empty image ref")
	}

	lastSlash := strings.LastIndex(trimmed, "/")
	lastColon := strings.LastIndex(trimmed, ":")
	if lastColon > lastSlash {
		name := strings.TrimSpace(trimmed[:lastColon])
		tag := strings.TrimSpace(trimmed[lastColon+1:])
		if name == "" || tag == "" {
			return "", "", fmt.Errorf("invalid image ref %q", imageRef)
		}
		return name, tag, nil
	}

	return trimmed, "latest", nil
}

func normalizeCategory(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "web", "pwn", "reverse", "crypto", "misc", "forensics":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "misc"
	}
}

func normalizeDifficulty(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case model.ChallengeDifficultyBeginner:
		return model.ChallengeDifficultyBeginner
	case model.ChallengeDifficultyEasy:
		return model.ChallengeDifficultyEasy
	case model.ChallengeDifficultyMedium:
		return model.ChallengeDifficultyMedium
	case model.ChallengeDifficultyHard:
		return model.ChallengeDifficultyHard
	case model.ChallengeDifficultyInsane, "hell":
		return model.ChallengeDifficultyInsane
	default:
		return model.ChallengeDifficultyEasy
	}
}

func defaultPointsByDifficulty(difficulty string) int {
	switch difficulty {
	case model.ChallengeDifficultyBeginner:
		return 50
	case model.ChallengeDifficultyEasy:
		return 100
	case model.ChallengeDifficultyMedium:
		return 200
	case model.ChallengeDifficultyHard:
		return 300
	case model.ChallengeDifficultyInsane:
		return 500
	default:
		return 100
	}
}

func sanitizeFlagToken(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "challenge"
	}
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	result := strings.Trim(b.String(), "_-")
	if result == "" {
		return "challenge"
	}
	return result
}

func safeJoin(baseDir, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" {
		return "", fmt.Errorf("relative path is empty")
	}
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	target := filepath.Clean(filepath.Join(baseAbs, rel))
	if target == baseAbs {
		return target, nil
	}
	prefix := baseAbs + string(os.PathSeparator)
	if !strings.HasPrefix(target, prefix) {
		return "", fmt.Errorf("resolved path escapes pack dir: %s", rel)
	}
	return target, nil
}

func getenvBool(key string, defaultValue bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultValue
	}
	return value
}
