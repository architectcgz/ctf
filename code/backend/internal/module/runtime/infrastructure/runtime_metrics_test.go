package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/docker/docker/api/types"

	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func TestCollectManagedContainerStatsSkipsFailedContainers(t *testing.T) {
	t.Parallel()

	containers := []types.Container{
		{ID: "aaaaaaaaaaaa1111", Names: []string{"/web"}},
		{ID: "bbbbbbbbbbbb2222", Names: []string{"/db"}},
	}

	stats := collectManagedContainerStats(context.Background(), containers, func(_ context.Context, container types.Container) (runtimeports.ManagedContainerStat, error) {
		if container.ID == "bbbbbbbbbbbb2222" {
			return runtimeports.ManagedContainerStat{}, errors.New("stats unavailable")
		}
		return runtimeports.ManagedContainerStat{
			ContainerID:   shortContainerID(container.ID),
			ContainerName: "web",
			CPUPercent:    12.5,
		}, nil
	})

	if len(stats) != 1 {
		t.Fatalf("expected 1 successful stat, got %+v", stats)
	}
	if stats[0].ContainerID != "aaaaaaaaaaaa" || stats[0].ContainerName != "web" {
		t.Fatalf("unexpected stats result: %+v", stats)
	}
}

func TestCollectManagedContainerStatsPreservesSuccessfulOrder(t *testing.T) {
	t.Parallel()

	containers := []types.Container{
		{ID: "cccccccccccc3333", Names: []string{"/first"}},
		{ID: "dddddddddddd4444", Names: []string{"/second"}},
		{ID: "eeeeeeeeeeee5555", Names: []string{"/third"}},
	}

	stats := collectManagedContainerStats(context.Background(), containers, func(_ context.Context, container types.Container) (runtimeports.ManagedContainerStat, error) {
		switch container.ID {
		case "cccccccccccc3333":
			time.Sleep(20 * time.Millisecond)
			return runtimeports.ManagedContainerStat{ContainerID: "cccccccccccc", ContainerName: "first"}, nil
		case "dddddddddddd4444":
			return runtimeports.ManagedContainerStat{}, errors.New("decode failed")
		default:
			time.Sleep(5 * time.Millisecond)
			return runtimeports.ManagedContainerStat{ContainerID: "eeeeeeeeeeee", ContainerName: "third"}, nil
		}
	})

	if len(stats) != 2 {
		t.Fatalf("expected 2 successful stats, got %+v", stats)
	}
	if stats[0].ContainerName != "first" || stats[1].ContainerName != "third" {
		t.Fatalf("expected successful stats to preserve source order, got %+v", stats)
	}
}
