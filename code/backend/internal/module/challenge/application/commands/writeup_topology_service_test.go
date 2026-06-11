package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/apperror"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
	"ctf-platform/internal/shared/taxonomy"
)

func TestWriteupServiceUpsertAndGetPublished(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&challengeentity.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: challengeentity.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challengeItem := &challengeentity.Challenge{
		Title:       "web-101",
		Description: "desc",
		Category:    taxonomy.DimensionWeb,
		Difficulty:  challengeentity.ChallengeDifficultyEasy,
		Points:      100,
		ImageID:     int64Ptr(1),
		Status:      challengeentity.ChallengeStatusPublished,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	writeupRepo := challengeinfra.NewWriteupServiceRepository(repo)
	service := NewWriteupService(writeupRepo)

	saved, err := service.Upsert(context.Background(), challengeItem.ID, 99, UpsertOfficialWriteupInput{
		Title:      "官方题解",
		Content:    "## Step 1",
		Visibility: challengeentity.WriteupVisibilityPublic,
	})
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if saved.Title != "官方题解" {
		t.Fatalf("unexpected writeup title: %+v", saved)
	}

	queryService := challengeqry.NewWriteupService(writeupRepo)
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
	if err := db.Create(&challengeentity.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: challengeentity.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image 1: %v", err)
	}
	if err := db.Create(&challengeentity.Image{ID: 2, Name: "ctf/db", Tag: "v1", Status: challengeentity.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image 2: %v", err)
	}
	challengeItem := &challengeentity.Challenge{
		Title:       "web-201",
		Description: "desc",
		Category:    taxonomy.DimensionWeb,
		Difficulty:  challengeentity.ChallengeDifficultyMedium,
		Points:      200,
		ImageID:     int64Ptr(1),
		Status:      challengeentity.ChallengeStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	templateRepo := challengeinfra.NewTemplateRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := NewTopologyService(
		challengeinfra.NewTopologyServiceRepository(repo),
		challengeinfra.NewTopologyTemplateRepository(templateRepo),
		challengeinfra.NewImageQueryRepository(imageRepo),
	)

	templateResp, err := service.CreateTemplate(context.Background(), UpsertEnvironmentTemplateInput{
		Name:         "双节点模板",
		Description:  "web + db",
		EntryNodeKey: "web",
		Networks: []challengecontracts.TopologyNetworkReq{
			{Key: "public", Name: "Public"},
			{Key: "backend", Name: "Backend", Internal: true},
		},
		Nodes: []challengecontracts.TopologyNodeReq{
			{Key: "web", Name: "Web", ImageID: 1, ServicePort: 8080, Tier: challengecontracts.TopologyTierPublic, NetworkKeys: []string{"public", "backend"}},
			{Key: "db", Name: "DB", ImageID: 2, Tier: challengecontracts.TopologyTierInternal, NetworkKeys: []string{"backend"}},
		},
		Links: []challengecontracts.TopologyLinkReq{
			{FromNodeKey: "web", ToNodeKey: "db"},
		},
		Policies: []challengecontracts.TopologyTrafficPolicyReq{
			{SourceNodeKey: "web", TargetNodeKey: "db", Action: challengecontracts.TopologyPolicyActionAllow},
			{SourceNodeKey: "db", TargetNodeKey: "web", Action: challengecontracts.TopologyPolicyActionDeny},
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

	queryService := challengeqry.NewTopologyService(
		challengeinfra.NewTopologyServiceRepository(repo),
		challengeinfra.NewTopologyTemplateRepository(templateRepo),
	)
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
	if err := db.Create(&challengeentity.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: challengeentity.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challengeItem := &challengeentity.Challenge{
		Title:       "web-202",
		Description: "desc",
		Category:    taxonomy.DimensionWeb,
		Difficulty:  challengeentity.ChallengeDifficultyMedium,
		Points:      200,
		ImageID:     int64Ptr(1),
		Status:      challengeentity.ChallengeStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := NewTopologyService(
		challengeinfra.NewTopologyServiceRepository(challengeinfra.NewRepository(db)),
		challengeinfra.NewTopologyTemplateRepository(challengeinfra.NewTemplateRepository(db)),
		challengeinfra.NewImageQueryRepository(challengeinfra.NewImageRepository(db)),
	)
	_, err := service.SaveChallengeTopology(context.Background(), challengeItem.ID, SaveChallengeTopologyInput{
		EntryNodeKey: "web",
		Networks: []challengecontracts.TopologyNetworkReq{
			{Key: "public", Name: "Public"},
		},
		Nodes: []challengecontracts.TopologyNodeReq{
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
	if err := db.Create(&challengeentity.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: challengeentity.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challengeItem := &challengeentity.Challenge{
		Title:           "shared-web-flag",
		Description:     "desc",
		Category:        taxonomy.DimensionWeb,
		Difficulty:      challengeentity.ChallengeDifficultyMedium,
		Points:          200,
		ImageID:         int64Ptr(1),
		Status:          challengeentity.ChallengeStatusDraft,
		InstanceSharing: challengeentity.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := NewTopologyService(
		challengeinfra.NewTopologyServiceRepository(challengeinfra.NewRepository(db)),
		challengeinfra.NewTopologyTemplateRepository(challengeinfra.NewTemplateRepository(db)),
		challengeinfra.NewImageQueryRepository(challengeinfra.NewImageRepository(db)),
	)
	_, err := service.SaveChallengeTopology(context.Background(), challengeItem.ID, SaveChallengeTopologyInput{
		EntryNodeKey: "web",
		Nodes: []challengecontracts.TopologyNodeReq{
			{Key: "web", Name: "Web", ImageID: 1, ServicePort: 8080, InjectFlag: true},
		},
	})
	if err == nil {
		t.Fatalf("expected shared challenge topology validation error")
	}
	if err.Error() != apperror.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for shared inject_flag topology, got %v", err)
	}
}

func TestTopologyServiceAllowsFineGrainedPolicyOnTemplateCreate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&challengeentity.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: challengeentity.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := NewTopologyService(
		challengeinfra.NewTopologyServiceRepository(challengeinfra.NewRepository(db)),
		challengeinfra.NewTopologyTemplateRepository(challengeinfra.NewTemplateRepository(db)),
		challengeinfra.NewImageQueryRepository(challengeinfra.NewImageRepository(db)),
	)
	saved, err := service.CreateTemplate(context.Background(), UpsertEnvironmentTemplateInput{
		Name:         "细粒度策略模板",
		EntryNodeKey: "web",
		Nodes: []challengecontracts.TopologyNodeReq{
			{Key: "web", Name: "Web", ImageID: 1, ServicePort: 8080},
		},
		Policies: []challengecontracts.TopologyTrafficPolicyReq{
			{SourceNodeKey: "web", TargetNodeKey: "web", Action: challengecontracts.TopologyPolicyActionAllow, Protocol: challengecontracts.TopologyPolicyProtocolTCP, Ports: []int{8080}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}
	if len(saved.Policies) != 1 || saved.Policies[0].Protocol != challengecontracts.TopologyPolicyProtocolTCP {
		t.Fatalf("unexpected fine-grained policy payload: %+v", saved.Policies)
	}
}

func TestTopologyServiceAllowsFineGrainedPolicyWhenBindingTemplate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&challengeentity.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: challengeentity.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challengeItem := &challengeentity.Challenge{
		Title:       "web-203",
		Description: "desc",
		Category:    taxonomy.DimensionWeb,
		Difficulty:  challengeentity.ChallengeDifficultyMedium,
		Points:      200,
		ImageID:     int64Ptr(1),
		Status:      challengeentity.ChallengeStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	rawSpec, err := challengecontracts.EncodeTopologySpec(challengecontracts.TopologySpec{
		Nodes: []challengecontracts.TopologyNode{
			{Key: "web", Name: "Web", ImageID: 1, ServicePort: 8080},
		},
		Policies: []challengecontracts.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "web", Action: challengecontracts.TopologyPolicyActionAllow, Protocol: challengecontracts.TopologyPolicyProtocolTCP, Ports: []int{8080}},
		},
	})
	if err != nil {
		t.Fatalf("encode spec: %v", err)
	}
	template := &challengeentity.EnvironmentTemplate{
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

	service := NewTopologyService(
		challengeinfra.NewTopologyServiceRepository(challengeinfra.NewRepository(db)),
		challengeinfra.NewTopologyTemplateRepository(challengeinfra.NewTemplateRepository(db)),
		challengeinfra.NewImageQueryRepository(challengeinfra.NewImageRepository(db)),
	)
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
