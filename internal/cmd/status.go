package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ejfox/dockrune/internal/config"
	"github.com/ejfox/dockrune/internal/storage"
	"github.com/spf13/cobra"
)

func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show deployment status",
		Long:  `Display the status of recent and active deployments`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}
}

func runStatus() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	store, err := storage.NewSQLiteStorage(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer store.Close()

	deployments, err := store.ListDeployments(10)
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(deployments) == 0 {
		fmt.Println("No deployments found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tREPO\tENV\tSTATUS\tURL")
	fmt.Fprintln(w, "--\t----\t---\t------\t---")

	for _, d := range deployments {
		fmt.Fprintf(w, "%s\t%s/%s\t%s\t%s\t%s\n",
			d.ID[:12],
			d.Owner,
			d.Repo,
			d.Environment,
			d.Status,
			d.URL,
		)
	}

	w.Flush()
	return nil
}
