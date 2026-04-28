package infrastructure_test

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ctf-platform/internal/model"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
)

func TestRepositoryReserveAvailablePortSkipsAllocatedPort(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.PortAllocation{}); err != nil {
		t.Fatalf("migrate port allocations: %v", err)
	}
	if err := db.Create(&model.PortAllocation{Port: 30000}).Error; err != nil {
		t.Fatalf("seed allocated port: %v", err)
	}

	repo := practiceinfra.NewRepository(db)
	port, err := repo.ReserveAvailablePort(context.Background(), 30000, 30002)
	if err != nil {
		t.Fatalf("ReserveAvailablePort() error = %v", err)
	}
	if port != 30001 {
		t.Fatalf("expected port 30001, got %d", port)
	}

	var count int64
	if err := db.Model(&model.PortAllocation{}).Where("port IN ?", []int{30000, 30001}).Count(&count).Error; err != nil {
		t.Fatalf("count allocated ports: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected two allocated ports, got %d", count)
	}
}
