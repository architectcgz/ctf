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

	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/infrastructure/postgres"
	"ctf-platform/internal/model"
	challengedomain "ctf-platform/internal/module/challenge/domain"
	"ctf-platform/pkg/crypto"
)

const (
	defaultGlobalFlagSecret = "dev-integration-secret-123456789"
	defaultPacksDir         = "../../docs/challenges/packs"
)

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
	Level   int
	Title   string
	Content string
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
		manifestPath := filepath.Join(packDir, "challenge.yml")
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
	parsed, err := challengedomain.ParseChallengePackageDir(packDir)
	if err != nil {
		return false, false, err
	}

	spec, err := buildChallengeSpec(parsed)
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
		err = findChallengeForPackageUpsert(tx, spec.Slug, spec.Title, spec.Category, &challenge)

		now := time.Now()
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			challenge = model.Challenge{
				PackageSlug:   stringPtr(spec.Slug),
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
				"package_slug":   spec.Slug,
				"title":          spec.Title,
				"description":    spec.Description,
				"category":       spec.Category,
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

func findChallengeForPackageUpsert(
	tx *gorm.DB,
	packageSlug string,
	title string,
	category string,
	challenge *model.Challenge,
) error {
	if challenge == nil {
		return fmt.Errorf("challenge target is nil")
	}

	slug := strings.TrimSpace(packageSlug)
	if slug != "" {
		err := tx.Unscoped().
			Where("package_slug = ?", slug).
			First(challenge).Error
		switch {
		case err == nil:
			return nil
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}
	}

	return tx.Unscoped().
		Where("(package_slug IS NULL OR package_slug = '') AND title = ? AND category = ?", title, category).
		First(challenge).Error
}

func buildChallengeSpec(parsed *challengedomain.ParsedChallengePackage) (*challengeSpec, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed challenge package is nil")
	}

	hints := make([]hintSpec, 0, len(parsed.Hints))
	for _, hint := range parsed.Hints {
		hints = append(hints, hintSpec{
			Level:   hint.Level,
			Title:   hint.Title,
			Content: hint.Content,
		})
	}

	attachmentURL := ""
	if len(parsed.Attachments) > 0 {
		attachmentURL = buildAttachmentURL(parsed.Slug, parsed.Attachments[0].Path)
	}

	return &challengeSpec{
		Slug:        parsed.Slug,
		Title:       parsed.Title,
		Description: parsed.Description,
		Category:    parsed.Category,
		Difficulty:  parsed.Difficulty,
		Points:      parsed.Points,
		ImageID:     0,
		ImageRef:    parsed.RuntimeImageRef,
		Attachment:  attachmentURL,
		Hints:       hints,
		FlagPrefix:  parsed.FlagPrefix,
		FlagValue:   parsed.FlagValue,
	}, nil
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

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
