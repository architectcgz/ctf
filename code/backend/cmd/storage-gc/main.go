package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/infrastructure/postgres"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "storage-gc: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("storage-gc", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	env := flags.String("env", envOrDefault("APP_ENV", "dev"), "config environment")
	kind := flags.String("kind", "files", "gc kind: files")
	execute := flags.Bool("execute", false, "delete eligible local files")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *kind != "files" {
		return fmt.Errorf("unsupported gc kind %q", *kind)
	}

	cfg, err := config.Load(*env)
	if err != nil {
		return err
	}
	db, err := postgres.Open(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	roots := challengeinfra.DefaultChallengeLocalStorageRootsFromEnv()
	service := challengecmd.NewArtifactGCService(
		challengecmd.DefaultArtifactGCConfig(time.Now().UTC(), challengecmd.ArtifactGCRoots{
			PreviewRoots: []string{
				roots.ChallengeImportPreviewRoot,
				roots.AWDChallengeImportPreviewRoot,
			},
			AttachmentRoot:         roots.ChallengeAttachmentRoot,
			ImageBuildSourceRoot:   roots.ImageBuildSourceRoot,
			AWDCheckerArtifactRoot: roots.AWDCheckerArtifactRoot,
		}),
		challengeinfra.NewArtifactReferenceRepository(db),
	)
	report, err := service.CollectFiles(ctx, challengecmd.ArtifactGCExecution{DryRun: !*execute})
	if err != nil {
		return err
	}
	mode := "dry-run"
	if *execute {
		mode = "execute"
	}
	fmt.Printf("storage-gc mode=%s kind=files candidates=%d deleted=%d\n", mode, len(report.Candidates), report.DeletedCount)
	for _, candidate := range report.Candidates {
		status := "delete"
		if candidate.Protected {
			status = "keep"
		}
		fmt.Printf("%s\t%s\t%s\t%d\t%s\n", status, candidate.Kind, candidate.Reason, candidate.SizeBytes, candidate.Path)
	}
	return nil
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
