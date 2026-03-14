package practice

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge"
)

func TestBuildTopologyCreateRequestKeepsFineGrainedPolicies(t *testing.T) {
	db := setupPracticeTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := &Service{
		imageRepo: challenge.NewImageRepository(db),
		config:    &config.Config{},
	}

	request, err := service.buildTopologyCreateRequest(&model.Challenge{ImageID: 1}, "web", model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
		Policies: []model.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "web", Action: model.TopologyPolicyActionAllow, Protocol: model.TopologyPolicyProtocolTCP, Ports: []int{8080}},
		},
	}, "flag{demo}")
	if err != nil {
		t.Fatalf("buildTopologyCreateRequest() error = %v", err)
	}
	if len(request.Policies) != 1 {
		t.Fatalf("expected fine-grained policy to be kept, got %+v", request.Policies)
	}
	if request.Policies[0].Protocol != model.TopologyPolicyProtocolTCP {
		t.Fatalf("unexpected policy protocol: %+v", request.Policies[0])
	}
}

func setupPracticeTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Image{}); err != nil {
		t.Fatalf("migrate image: %v", err)
	}
	return db
}
