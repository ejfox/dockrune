package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func DeployCmd() *cobra.Command {
	var owner, repo, ref string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Manually trigger a deployment",
		Long:  `Manually trigger a deployment for a specific repository and ref`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(owner, repo, ref)
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Repository owner")
	cmd.Flags().StringVar(&repo, "repo", "", "Repository name")
	cmd.Flags().StringVar(&ref, "ref", "main", "Git ref to deploy")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("repo")

	return cmd
}

func runDeploy(owner, repo, ref string) error {
	// This would connect to the running server and trigger a deployment
	fmt.Printf("Triggering deployment for %s/%s@%s\n", owner, repo, ref)
	fmt.Println("Note: This feature requires the server to be running")
	return nil
}
