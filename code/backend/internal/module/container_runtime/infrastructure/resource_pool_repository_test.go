package infrastructure

import (
	"context"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"ctf-platform/internal/config"
	runtimeentity "ctf-platform/internal/module/container_runtime/entity"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func newRuntimeResourcePoolRepositoryTestDB(t *testing.T) *gorm.DB {
	return newRuntimeResourcePoolRepositoryTestDBWithLogger(t, nil)
}

func newRuntimeResourcePoolRepositoryTestDBWithLogger(t *testing.T, logger gormlogger.Interface) *gorm.DB {
	t.Helper()

	cfg := &gorm.Config{}
	if logger != nil {
		cfg.Logger = logger
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), cfg)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(
		&runtimeentity.RuntimeNode{},
		&runtimeentity.RuntimePortPool{},
		&runtimeentity.RuntimeSubnetPool{},
	); err != nil {
		t.Fatalf("migrate runtime resource pool: %v", err)
	}
	return db
}

func TestResourcePoolAllowsDifferentNodesToReserveSamePort(t *testing.T) {
	t.Parallel()

	db := newRuntimeResourcePoolRepositoryTestDB(t)
	repo := NewRuntimeResourcePoolRepository(db)
	seedRuntimePortPoolRows(t, db,
		runtimeentity.RuntimePortPool{RuntimeNodeID: 101, Port: 30000, Status: runtimeentity.RuntimeResourceStatusAvailable},
		runtimeentity.RuntimePortPool{RuntimeNodeID: 202, Port: 30000, Status: runtimeentity.RuntimeResourceStatusAvailable},
	)

	nodeAPort, err := repo.ReserveAvailablePortForNode(context.Background(), 101, 1001)
	if err != nil {
		t.Fatalf("ReserveAvailablePortForNode(node A) error = %v", err)
	}
	nodeBPort, err := repo.ReserveAvailablePortForNode(context.Background(), 202, 2002)
	if err != nil {
		t.Fatalf("ReserveAvailablePortForNode(node B) error = %v", err)
	}
	if nodeAPort != 30000 || nodeBPort != 30000 {
		t.Fatalf("reserved ports = node A %d node B %d, want both 30000", nodeAPort, nodeBPort)
	}
}

func TestResourcePoolDoesNotReturnSamePortForSameNodeReservations(t *testing.T) {
	t.Parallel()

	db := newRuntimeResourcePoolRepositoryTestDB(t)
	repo := NewRuntimeResourcePoolRepository(db)
	seedRuntimePortPoolRows(t, db,
		runtimeentity.RuntimePortPool{RuntimeNodeID: 101, Port: 30000, Status: runtimeentity.RuntimeResourceStatusAvailable},
		runtimeentity.RuntimePortPool{RuntimeNodeID: 101, Port: 30001, Status: runtimeentity.RuntimeResourceStatusAvailable},
	)

	type reservationResult struct {
		port int
		err  error
	}
	results := make(chan reservationResult, 2)
	for _, instanceID := range []int64{1001, 1002} {
		instanceID := instanceID
		go func() {
			port, err := repo.ReserveAvailablePortForNode(context.Background(), 101, instanceID)
			results <- reservationResult{port: port, err: err}
		}()
	}

	ports := make([]int, 0, 2)
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("ReserveAvailablePortForNode() error = %v", result.err)
		}
		ports = append(ports, result.port)
	}
	sort.Ints(ports)
	if ports[0] != 30000 || ports[1] != 30001 {
		t.Fatalf("reserved ports = %v, want [30000 30001]", ports)
	}
}

func TestResourcePoolAllowsDifferentNodesToReserveSameSubnet(t *testing.T) {
	t.Parallel()

	db := newRuntimeResourcePoolRepositoryTestDB(t)
	repo := NewRuntimeResourcePoolRepository(db)
	seedRuntimeSubnetPoolRows(t, db,
		runtimeentity.RuntimeSubnetPool{RuntimeNodeID: 101, PoolKind: runtimeentity.RuntimeSubnetPoolKindTopology, Subnet: "10.10.0.0/24", Status: runtimeentity.RuntimeResourceStatusAvailable},
		runtimeentity.RuntimeSubnetPool{RuntimeNodeID: 202, PoolKind: runtimeentity.RuntimeSubnetPoolKindTopology, Subnet: "10.10.0.0/24", Status: runtimeentity.RuntimeResourceStatusAvailable},
	)

	nodeASubnet, err := repo.ReserveAvailableSubnetForNode(context.Background(), 101, runtimeentity.RuntimeSubnetPoolKindTopology, 1001, "topology")
	if err != nil {
		t.Fatalf("ReserveAvailableSubnetForNode(node A) error = %v", err)
	}
	nodeBSubnet, err := repo.ReserveAvailableSubnetForNode(context.Background(), 202, runtimeentity.RuntimeSubnetPoolKindTopology, 2002, "topology")
	if err != nil {
		t.Fatalf("ReserveAvailableSubnetForNode(node B) error = %v", err)
	}
	if nodeASubnet != "10.10.0.0/24" || nodeBSubnet != "10.10.0.0/24" {
		t.Fatalf("reserved subnets = node A %q node B %q, want both 10.10.0.0/24", nodeASubnet, nodeBSubnet)
	}
}

func TestResourcePoolEnsurePoolsForNodeIsIdempotentAndPreservesAllocatedRows(t *testing.T) {
	t.Parallel()

	db := newRuntimeResourcePoolRepositoryTestDB(t)
	repo := NewRuntimeResourcePoolRepository(db)
	instanceID := int64(1001)
	seedRuntimePortPoolRows(t, db,
		runtimeentity.RuntimePortPool{
			RuntimeNodeID: 101,
			Port:          30000,
			Status:        runtimeentity.RuntimeResourceStatusReserved,
			InstanceID:    &instanceID,
		},
	)
	seedRuntimeSubnetPoolRows(t, db,
		runtimeentity.RuntimeSubnetPool{
			RuntimeNodeID: 101,
			PoolKind:      runtimeentity.RuntimeSubnetPoolKindTopology,
			Subnet:        "10.10.0.0/24",
			Status:        runtimeentity.RuntimeResourceStatusBound,
			InstanceID:    &instanceID,
			NetworkKey:    "topology",
		},
	)

	cfg := config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30003,
		Network: config.ContainerNetworkConfig{
			SingleContainerSubnetBase: "10.11.0.0/29",
			SingleContainerSubnetMask: 30,
			TopologySubnetBase:        "10.10.0.0/16",
			TopologySubnetMask:        24,
		},
	}
	for range 2 {
		if err := repo.EnsurePoolsForNode(context.Background(), 101, cfg); err != nil {
			t.Fatalf("EnsurePoolsForNode() error = %v", err)
		}
	}

	if got := countRuntimePortPoolRows(t, db, 101); got != 3 {
		t.Fatalf("port pool rows = %d, want 3", got)
	}
	if got := countRuntimeSubnetPoolRows(t, db, 101, runtimeentity.RuntimeSubnetPoolKindTopology); got != 256 {
		t.Fatalf("topology subnet pool rows = %d, want 256", got)
	}
	if got := countRuntimeSubnetPoolRows(t, db, 101, runtimeentity.RuntimeSubnetPoolKindSingleContainer); got != 2 {
		t.Fatalf("single-container subnet pool rows = %d, want 2", got)
	}

	var reservedPort runtimeentity.RuntimePortPool
	if err := db.Where("runtime_node_id = ? AND port = ?", 101, 30000).First(&reservedPort).Error; err != nil {
		t.Fatalf("load reserved port: %v", err)
	}
	if reservedPort.Status != runtimeentity.RuntimeResourceStatusReserved || reservedPort.InstanceID == nil || *reservedPort.InstanceID != instanceID {
		t.Fatalf("reserved port row = %+v, want reserved for instance %d", reservedPort, instanceID)
	}

	var boundSubnet runtimeentity.RuntimeSubnetPool
	if err := db.Where("runtime_node_id = ? AND subnet = ?", 101, "10.10.0.0/24").First(&boundSubnet).Error; err != nil {
		t.Fatalf("load bound subnet: %v", err)
	}
	if boundSubnet.Status != runtimeentity.RuntimeResourceStatusBound || boundSubnet.InstanceID == nil || *boundSubnet.InstanceID != instanceID {
		t.Fatalf("bound subnet row = %+v, want bound for instance %d", boundSubnet, instanceID)
	}
}

func TestResourcePoolEnsurePoolsForNodeBatchesSeedInserts(t *testing.T) {
	t.Parallel()

	insertCounter := newRuntimeResourcePoolInsertCounter()
	db := newRuntimeResourcePoolRepositoryTestDBWithLogger(t, insertCounter)
	repo := NewRuntimeResourcePoolRepository(db)
	insertCounter.Reset()

	cfg := config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30020,
		Network: config.ContainerNetworkConfig{
			SingleContainerSubnetBase: "10.11.0.0/24",
			SingleContainerSubnetMask: 28,
			TopologySubnetBase:        "10.10.0.0/24",
			TopologySubnetMask:        28,
		},
	}
	if err := repo.EnsurePoolsForNode(context.Background(), 101, cfg); err != nil {
		t.Fatalf("EnsurePoolsForNode() error = %v", err)
	}

	portRows := countRuntimePortPoolRows(t, db, 101)
	subnetRows := countRuntimeSubnetPoolRows(t, db, 101, runtimeentity.RuntimeSubnetPoolKindSingleContainer) +
		countRuntimeSubnetPoolRows(t, db, 101, runtimeentity.RuntimeSubnetPoolKindTopology)
	portInsertStatements, subnetInsertStatements := insertCounter.InsertStatements()
	if portInsertStatements >= int(portRows) {
		t.Fatalf("runtime port pool insert statements = %d for %d rows, want batched inserts", portInsertStatements, portRows)
	}
	if subnetInsertStatements >= int(subnetRows) {
		t.Fatalf("runtime subnet pool insert statements = %d for %d rows, want batched inserts", subnetInsertStatements, subnetRows)
	}
}

func seedRuntimePortPoolRows(t *testing.T, db *gorm.DB, rows ...runtimeentity.RuntimePortPool) {
	t.Helper()
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed runtime port pool: %v", err)
	}
}

func countRuntimePortPoolRows(t *testing.T, db *gorm.DB, nodeID int64) int64 {
	t.Helper()

	var count int64
	if err := db.Model(&runtimeentity.RuntimePortPool{}).
		Where("runtime_node_id = ?", nodeID).
		Count(&count).Error; err != nil {
		t.Fatalf("count runtime port pool: %v", err)
	}
	return count
}

func countRuntimeSubnetPoolRows(t *testing.T, db *gorm.DB, nodeID int64, poolKind string) int64 {
	t.Helper()

	var count int64
	if err := db.Model(&runtimeentity.RuntimeSubnetPool{}).
		Where("runtime_node_id = ? AND pool_kind = ?", nodeID, poolKind).
		Count(&count).Error; err != nil {
		t.Fatalf("count runtime subnet pool: %v", err)
	}
	return count
}

func seedRuntimeSubnetPoolRows(t *testing.T, db *gorm.DB, rows ...runtimeentity.RuntimeSubnetPool) {
	t.Helper()
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed runtime subnet pool: %v", err)
	}
}

type runtimeResourcePoolInsertCounter struct {
	gormlogger.Interface

	mu                     sync.Mutex
	portInsertStatements   int
	subnetInsertStatements int
}

func newRuntimeResourcePoolInsertCounter() *runtimeResourcePoolInsertCounter {
	return &runtimeResourcePoolInsertCounter{
		Interface: gormlogger.Default.LogMode(gormlogger.Silent),
	}
}

func (l *runtimeResourcePoolInsertCounter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.portInsertStatements = 0
	l.subnetInsertStatements = 0
}

func (l *runtimeResourcePoolInsertCounter) InsertStatements() (int, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.portInsertStatements, l.subnetInsertStatements
}

func (l *runtimeResourcePoolInsertCounter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	l.Interface = l.Interface.LogMode(level)
	return l
}

func (l *runtimeResourcePoolInsertCounter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	normalized := strings.ToLower(sql)
	l.mu.Lock()
	switch {
	case strings.Contains(normalized, "insert into `runtime_port_pool`") ||
		strings.Contains(normalized, `insert into "runtime_port_pool"`) ||
		strings.Contains(normalized, "insert into runtime_port_pool"):
		l.portInsertStatements++
	case strings.Contains(normalized, "insert into `runtime_subnet_pool`") ||
		strings.Contains(normalized, `insert into "runtime_subnet_pool"`) ||
		strings.Contains(normalized, "insert into runtime_subnet_pool"):
		l.subnetInsertStatements++
	}
	l.mu.Unlock()
	l.Interface.Trace(ctx, begin, func() (string, int64) {
		return sql, rows
	}, err)
}
