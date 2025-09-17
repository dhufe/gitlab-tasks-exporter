package main

import (
	"fmt"
	"os"

	"hufschlaeger.net/gitlab-tasks-exporter/internal/cli"
	"hufschlaeger.net/gitlab-tasks-exporter/internal/service"
)

func main() {
	cfg, err := cli.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Fehler beim Parsen der Flags: %v\n", err)
		os.Exit(1)
	}

	exporter := service.NewExporter(cfg)

	if err := exporter.Export(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Export fehlgeschlagen: %v\n", err)
		os.Exit(1)
	}
}
