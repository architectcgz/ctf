package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
	"ctf-platform/pkg/errcode"
)

func TestWriteupServiceUpsertAndGetPublished(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challengeItem := &model.Challenge{
		Title:       "web-101",
		Description: "desc",
		Category:    model.DimensionWeb,
		Difficulty:  model.ChallengeDifficultyEasy,
		Points:      100,
		ImageID:     1,
		Status:      model.ChallengeStatusPublished,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	service := NewWriteupService(repo)

	saved, err := service.Upsert(context.Background(), challengeItem.ID, 99, UpsertOfficialWriteupInput{
		Title:      "官方题解",
		Content:    "## Step 1",
		Visibility: model.WriteupVisibilityPublic,
	})
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if saved.Title != "官方题解" {
		t.Fatalf("unexpected writeup title: %+v", saved)
	}

	queryService := challengeqry.NewWriteupService(repo)
	published, err := queryService.GetPublished(context.Background(), 1001, challengeItem.ID)
	if err != nil {
		t.Fatalf("GetPublished() error = %v", err)
	}
	if !published.RequiresSpoilerWarning {
		t.Fatalf("unexpected published writeup: %+v", published)
	}
}

func TestTopologyServiceSaveChallengeTopologyWithTemplate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image 1: %v", err)
	}
	if err := db.Create(&model.Image{ID: 2, Name: "ctf/db", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image 2: %v", err)
	}
	challengeItem := &model.Challenge{
		Title:       "web-201",
		Description: "desc",
		Category:    model.DimensionWeb,
		Difficulty:  model.ChallengeDifficultyMedium,
		Points:      200,
		ImageID:     1,
		Status:      model.ChallengeStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	templateRepo := challengeinfra.NewTemplateRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := NewTopologyService(repo, templateRepo, imageRepo)

	templateResp, err := service.CreateTemplate(context.Background(), UpsertEnvironmentTemplateInput{
		Name:         "双节点模板",
		Description:  "web + db",
		EntryNodeKey: "web",
		Networks: []dto.TopologyNetworkReq{
			{Key: "public", Name: "Public"},
			{Key: "backend", Name: "Backend", Internal: true},
		},
		Nodes: []dto.TopologyNodeReq{
			{Key: "web", Name: "Web", ImageID: 1, ServicePort: 8080, Tier: model.TopologyTierPublic, NetworkKeys: []string{"public", "backend"}},
			{Key: "db", Name: "DB", ImageID: 2, Tier: model.TopologyTierInternal, NetworkKeys: []string{"backend"}},
		},
		Links: []dto.TopologyLinkReq{
			{FromNodeKey: "web", ToNodeKey: "db"},
		},
		Policies: []dto.TopologyTrafficPolicyReq{
			{SourceNodeKey: "web", TargetNodeKey: "db", Action: model.TopologyPolicyActionAllow},
			{SourceNodeKey: "db", TargetNodeKey: "web", Action: model.TopologyPolicyActionDeny},
		},
	})
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}

	saved, err := service.SaveChallengeTopology(context.Background(), challengeItem.ID, SaveChallengeTopologyInput{
		TemplateID: &templateResp.ID,
	})
	if err != nil {
		t.Fatalf("SaveChallengeTopology() error = %v", err)
	}
	if saved.TemplateID == nil || *saved.TemplateID != templateResp.ID {
		t.Fatalf("unexpected topology template binding: %+v", saved)
	}
	if len(saved.Nodes) != 2 || saved.EntryNodeKey != "web" {
		t.Fatalf("unexpected topology response: %+v", saved)
	}
	if len(saved.Networks) != 2 || len(saved.Policies) != 2 {
		t.Fatalf("unexpected topology segmentation response: %+v", saved)
	}
	if got := saved.Nodes[0].NetworkKeys; len(got) != 2 {
		t.Fatalf("unexpected node network keys: %+v", saved.Nodes[0])
	}

	queryService := challengeqry.NewTopologyService(repo, templateRepo)
	loadedTemplate, err := queryService.GetTemplate(context.Background(), templateResp.ID)
	if err != nil {
		t.Fatalf("GetTemplate() error = %v", err)
	}
	if loadedTemplate.UsageCount != 1 {
		t.Fatalf("expected usage_count=1, got %d", loadedTemplate.UsageCount)
	}
	if len(loadedTemplate.Networks) != 2 || len(loadedTemplate.Policies) != 2 {
		t.Fatalf("unexpected loaded template topology: %+v", loadedTemplate)
	}
}

func TestTopologyServiceRejectsUnknownNetworkReference(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challengeItem := &model.Challenge{
		Title:       "web-202",
		Description: "desc",
		Category:    model.DimensionWeb,
		Difficulty:  model.ChallengeDifficultyMedium,
		Points:      200,
		ImageID:     1,
		Status:      model.ChallengeStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := NewTopologyService(challengeinfra.NewRepository(db), challengeinfra.NewTemplateRepository(db), challengeinfra.NewImageRepository(db))
	_, err := service.SaveChallengeTopology(context.Background(), challengeItem.ID, SaveChallengeTopologyInput{
		EntryNodeKey: "web",
		Networks: []dto.TopologyNetworkReq{
			{Key: "public", Name: "Public"},
		},
		Nodes: []dto.TopologyNodeReq{
			{Key: "web", Name: "Web", ImageID: 1, ServicePort: 8080, NetworkKeys: []string{"missing"}},
		},
	})
	if err == nil {
		t.Fatalf("expected unknown network validation error")
	}
}

func TestTopologyServiceRejectsInjectFlagForSharedChallenge(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challengeItem := &model.Challenge{
		Title:           "shared-web-flag",
		Description:     "desc",
		Category:        model.DimensionWeb,
		Difficulty:      model.ChallengeDifficultyMedium,
		Points:          200,
		ImageID:         1,
		Status:          model.ChallengeStatusDraft,
		InstanceSharing: model.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := NewTopologyService(challengeinfra.NewRepository(db), challengeinfra.NewTemplateRepository(db), challengeinfra.NewImageRepository(db))
	_, err := service.SaveChallengeTopology(context.Background(), challengeItem.ID, SaveChallengeTopologyInput{
		EntryNodeKey: "web",
		Nodes: []dto.TopologyNodeReq{
			{Key: "web", Name: "Web", ImageID: 1, ServicePort: 8080, InjectFlag: true},
		},
	})
	if err == nil {
		t.Fatalf("expected shared challenge topology validation error")
	}
	if err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for shared inject_flag topology, got %v", err)
	}
}

func TestTopologyServiceAllowsFineGrainedPolicyOnTemplateCreate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := NewTopologyService(challengeinfra.NewRepository(db), challengeinfra.NewTemplateRepository(db), challengeinfra.NewImageRepository(db))
	saved, err := service.CreateTemplate(context.Background(), UpsertEnvironmentTemplateInput{
		Name:         "细粒度策略模板",
		EntryNodeKey: "web",
		Nodes: []dto.TopologyNodeReq{
			{Key: "web", Name: "Web", ImageID: 1, ServicePort: 8080},
		},
		Policies: []dto.TopologyTrafficPolicyReq{
			{SourceNodeKey: "web", TargetNodeKey: "web", Action: model.TopologyPolicyActionAllow, Protocol: model.TopologyPolicyProtocolTCP, Ports: []int{8080}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}
	if len(saved.Policies) != 1 || saved.Policies[0].Protocol != model.TopologyPolicyProtocolTCP {
		t.Fatalf("unexpected fine-grained policy payload: %+v", saved.Policies)
	}
}

func TestTopologyServiceAllowsFineGrainedPolicyWhenBindingTemplate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challengeItem := &model.Challenge{
		Title:       "web-203",
		Description: "desc",
		Category:    model.DimensionWeb,
		Difficulty:  model.ChallengeDifficultyMedium,
		Points:      200,
		ImageID:     1,
		Status:      model.ChallengeStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	rawSpec, err := model.EncodeTopologySpec(model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", Name: "Web", ImageID: 1, ServicePort: 8080},
		},
		Policies: []model.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "web", Action: model.TopologyPolicyActionAllow, Protocol: model.TopologyPolicyProtocolTCP, Ports: []int{8080}},
		},
	})
	if err != nil {
		t.Fatalf("encode spec: %v", err)
	}
	template := &model.EnvironmentTemplate{
		Name:         "legacy-template",
		Description:  "legacy",
		EntryNodeKey: "web",
		Spec:         rawSpec,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}

	service := NewTopologyService(challengeinfra.NewRepository(db), challengeinfra.NewTemplateRepository(db), challengeinfra.NewImageRepository(db))
	saved, err := service.SaveChallengeTopology(context.Background(), challengeItem.ID, SaveChallengeTopologyInput{
		TemplateID: &template.ID,
	})
	if err != nil {
		t.Fatalf("SaveChallengeTopology() error = %v", err)
	}
	if len(saved.Policies) != 1 || len(saved.Policies[0].Ports) != 1 || saved.Policies[0].Ports[0] != 8080 {
		t.Fatalf("unexpected bound fine-grained policy: %+v", saved.Policies)
	}
}
