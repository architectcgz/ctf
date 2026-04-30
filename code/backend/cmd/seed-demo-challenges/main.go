package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/infrastructure/postgres"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
)

type seedImageSpec struct {
	Name        string
	Tag         string
	Description string
}

type seedHintSpec struct {
	Level   int
	Title   string
	Content string
}

type seedChallengeSpec struct {
	Title         string
	Description   string
	Category      string
	Difficulty    string
	Points        int
	ImageRef      string
	AttachmentURL string
	FlagType      string
	StaticFlag    string
	FlagPrefix    string
	Hints         []seedHintSpec
}

type seedResult struct {
	ImagesCreated     int
	ImagesUpdated     int
	ChallengesCreated int
	ChallengesUpdated int
}

func main() {
	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if env == "" {
		env = "dev"
	}
	if strings.TrimSpace(os.Getenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET")) == "" {
		if err := os.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "dev-integration-secret-123456789"); err != nil {
			panic(fmt.Errorf("set default CTF_CONTAINER_FLAG_GLOBAL_SECRET: %w", err))
		}
	}

	cfg, err := config.Load(env)
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	ctx := context.Background()
	db, err := postgres.Open(ctx, cfg.Postgres)
	if err != nil {
		panic(fmt.Errorf("open postgres: %w", err))
	}

	dockerClient, dockerErr := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if dockerErr != nil {
		fmt.Fprintf(os.Stderr, "warn: docker client unavailable: %v\n", dockerErr)
		dockerClient = nil
	}

	result, err := seedDemoChallenges(ctx, db, dockerClient)
	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"seed completed: images created=%d updated=%d, challenges created=%d updated=%d\n",
		result.ImagesCreated,
		result.ImagesUpdated,
		result.ChallengesCreated,
		result.ChallengesUpdated,
	)
}

func seedDemoChallenges(ctx context.Context, db *gorm.DB, dockerClient *client.Client) (*seedResult, error) {
	imageSpecs := []seedImageSpec{
		{
			Name:        "localhost:5000/ctf/hello-web",
			Tag:         "v20260106125101",
			Description: "本地预置静态 Web 热身题镜像，适合验证实例创建与基础提交流程。",
		},
		{
			Name:        "localhost:5000/ctf/find-the-secret",
			Tag:         "v20260106102757",
			Description: "本地预置静态 Web 搜索题镜像，适合做页面源码检索热身。",
		},
		{
			Name:        "ctf-web-sqli",
			Tag:         "latest",
			Description: "本地预置 SQL 注入练习镜像，管理员登录成功后会展示动态 Flag。",
		},
	}

	challengeSpecs := []seedChallengeSpec{
		{
			Title:       "Hello Web",
			Description: "启动实例后访问首页，阅读页面源码并提交页面中直接出现的欢迎 Flag。该题用于验证题目列表、实例启动和 Flag 提交流程。",
			Category:    "web",
			Difficulty:  model.ChallengeDifficultyBeginner,
			Points:      50,
			ImageRef:    "localhost:5000/ctf/hello-web:v20260106125101",
			FlagType:    model.FlagTypeStatic,
			StaticFlag:  "flag{hello_web_test_2024}",
			FlagPrefix:  "flag",
			Hints: []seedHintSpec{
				{Level: 1, Title: "先看源码", Content: "页面首屏已经足够接近答案，先查看 HTML 源码。"},
				{Level: 2, Title: "Flag 位置", Content: "Flag 直接写在首页 HTML 中，不需要额外爆破或登录。"},
			},
		},
		{
			Title:       "Find The Secret",
			Description: "这是一个静态页面检索热身题。启动实例后访问首页，定位隐藏的 Secret Flag 并提交。",
			Category:    "web",
			Difficulty:  model.ChallengeDifficultyEasy,
			Points:      100,
			ImageRef:    "localhost:5000/ctf/find-the-secret:v20260106102757",
			FlagType:    model.FlagTypeStatic,
			StaticFlag:  "flag{auto_build_test_success_2024}",
			FlagPrefix:  "flag",
			Hints: []seedHintSpec{
				{Level: 1, Title: "静态资源", Content: "先看首页源码，再考虑是否需要继续查看静态资源文件。"},
				{Level: 2, Title: "无需登录", Content: "这题不需要账号系统，答案已经在页面可访问内容中。"},
			},
		},
		{
			Title:       "SQL Injection Login Bypass",
			Description: "这是一个经典 SQL 注入登录绕过练习。目标是以管理员身份登录，进入后台后读取系统展示的 Flag。该题实例会按平台规则注入动态 Flag。",
			Category:    "web",
			Difficulty:  model.ChallengeDifficultyEasy,
			Points:      150,
			ImageRef:    "ctf-web-sqli:latest",
			FlagType:    model.FlagTypeDynamic,
			FlagPrefix:  "flag",
			Hints: []seedHintSpec{
				{Level: 1, Title: "登录逻辑", Content: "观察登录表单提交目标，留意后端是否直接拼接 SQL。"},
				{Level: 2, Title: "绕过思路", Content: "目标不是拿到密码，而是让条件表达式恒为真，尝试使用经典注入 payload。"},
			},
		},
	}

	result := &seedResult{}
	imageIDs := make(map[string]int64, len(imageSpecs))

	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, spec := range imageSpecs {
			imageID, created, err := upsertImage(ctx, tx, dockerClient, spec)
			if err != nil {
				return err
			}
			if created {
				result.ImagesCreated++
			} else {
				result.ImagesUpdated++
			}
			imageIDs[fmt.Sprintf("%s:%s", spec.Name, spec.Tag)] = imageID
		}

		for _, spec := range challengeSpecs {
			imageID, ok := imageIDs[spec.ImageRef]
			if !ok {
				return fmt.Errorf("seed image not found for challenge %s: %s", spec.Title, spec.ImageRef)
			}

			created, err := upsertChallenge(tx, spec, imageID)
			if err != nil {
				return err
			}
			if created {
				result.ChallengesCreated++
			} else {
				result.ChallengesUpdated++
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func upsertImage(ctx context.Context, tx *gorm.DB, dockerClient *client.Client, spec seedImageSpec) (int64, bool, error) {
	size, err := inspectDockerImageSize(ctx, dockerClient, fmt.Sprintf("%s:%s", spec.Name, spec.Tag))
	if err != nil {
		return 0, false, err
	}

	var image model.Image
	err = tx.Unscoped().
		Where("name = ? AND tag = ?", spec.Name, spec.Tag).
		First(&image).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		image = model.Image{
			Name:        spec.Name,
			Tag:         spec.Tag,
			Description: spec.Description,
			Size:        size,
			Status:      model.ImageStatusAvailable,
		}
		if err := tx.Create(&image).Error; err != nil {
			return 0, false, fmt.Errorf("create image %s:%s: %w", spec.Name, spec.Tag, err)
		}
		return image.ID, true, nil
	case err != nil:
		return 0, false, fmt.Errorf("find image %s:%s: %w", spec.Name, spec.Tag, err)
	default:
		updates := map[string]any{
			"description": spec.Description,
			"size":        size,
			"status":      model.ImageStatusAvailable,
			"deleted_at":  nil,
			"updated_at":  time.Now(),
		}
		if err := tx.Model(&image).Updates(updates).Error; err != nil {
			return 0, false, fmt.Errorf("update image %s:%s: %w", spec.Name, spec.Tag, err)
		}
		return image.ID, false, nil
	}
}

func upsertChallenge(tx *gorm.DB, spec seedChallengeSpec, imageID int64) (bool, error) {
	var challenge model.Challenge
	err := tx.Unscoped().
		Where("title = ?", spec.Title).
		First(&challenge).Error

	now := time.Now()
	created := false
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		challenge = model.Challenge{
			Title:         spec.Title,
			Description:   spec.Description,
			Category:      spec.Category,
			Difficulty:    spec.Difficulty,
			Points:        spec.Points,
			ImageID:       imageID,
			AttachmentURL: spec.AttachmentURL,
			Status:        model.ChallengeStatusDraft,
			FlagPrefix:    spec.FlagPrefix,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Create(&challenge).Error; err != nil {
			return false, fmt.Errorf("create challenge %s: %w", spec.Title, err)
		}
		created = true
	case err != nil:
		return false, fmt.Errorf("find challenge %s: %w", spec.Title, err)
	default:
		updates := map[string]any{
			"description":    spec.Description,
			"category":       spec.Category,
			"difficulty":     spec.Difficulty,
			"points":         spec.Points,
			"image_id":       imageID,
			"attachment_url": spec.AttachmentURL,
			"flag_prefix":    spec.FlagPrefix,
			"deleted_at":     nil,
			"updated_at":     now,
		}
		if err := tx.Model(&challenge).Updates(updates).Error; err != nil {
			return false, fmt.Errorf("update challenge %s: %w", spec.Title, err)
		}
	}

	if err := syncChallengeHints(tx, challenge.ID, spec.Hints); err != nil {
		return false, err
	}

	if err := configureChallengeFlag(tx, challenge.ID, spec); err != nil {
		return false, err
	}

	if err := tx.Model(&model.Challenge{}).
		Where("id = ?", challenge.ID).
		Updates(map[string]any{
			"status":     model.ChallengeStatusPublished,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return false, fmt.Errorf("publish challenge %s: %w", spec.Title, err)
	}

	return created, nil
}

func syncChallengeHints(tx *gorm.DB, challengeID int64, hints []seedHintSpec) error {
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

func configureChallengeFlag(tx *gorm.DB, challengeID int64, spec seedChallengeSpec) error {
	updates := map[string]any{
		"flag_prefix": spec.FlagPrefix,
		"updated_at":  time.Now(),
	}

	switch spec.FlagType {
	case model.FlagTypeStatic:
		salt, err := crypto.GenerateSalt()
		if err != nil {
			return fmt.Errorf("generate salt for challenge %d: %w", challengeID, err)
		}
		updates["flag_type"] = model.FlagTypeStatic
		updates["flag_salt"] = salt
		updates["flag_hash"] = crypto.HashStaticFlag(spec.StaticFlag, salt)
	case model.FlagTypeDynamic:
		updates["flag_type"] = model.FlagTypeDynamic
		updates["flag_salt"] = ""
		updates["flag_hash"] = ""
	default:
		return fmt.Errorf("unsupported flag type for challenge %d: %s", challengeID, spec.FlagType)
	}

	if err := tx.Model(&model.Challenge{}).Where("id = ?", challengeID).Updates(updates).Error; err != nil {
		return fmt.Errorf("configure flag for challenge %d: %w", challengeID, err)
	}
	return nil
}

func inspectDockerImageSize(ctx context.Context, dockerClient *client.Client, imageRef string) (int64, error) {
	if dockerClient == nil {
		return 0, fmt.Errorf("docker client unavailable for image %s", imageRef)
	}

	inspectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	inspect, _, err := dockerClient.ImageInspectWithRaw(inspectCtx, imageRef)
	if err != nil {
		return 0, fmt.Errorf("inspect docker image %s: %w", imageRef, err)
	}
	return inspect.Size, nil
}
