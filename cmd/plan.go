package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cqi/my_agentmux/internal/workflow"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Manage workflow plans",
	Long: `Create, review, and manage spec-driven workflow plans.

Plans go through a lifecycle: draft → approved or rejected.
Use subcommands to create, list, approve, reject, show, or delete plans.`,
}

var planCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		description, _ := cmd.Flags().GetString("description")

		store, err := workflow.NewPlanStore(cfg.PlansDir())
		if err != nil {
			return err
		}

		plan, err := store.Create(title, description)
		if err != nil {
			return err
		}

		fmt.Printf("✓ Plan %s created: %s\n", plan.ID, plan.Title)
		if plan.Description != "" {
			fmt.Printf("  Description: %s\n", plan.Description)
		}
		fmt.Printf("  Status: %s\n", workflow.FormatStatus(plan.Status))
		fmt.Printf("  Approve: agentmux plan approve %s\n", plan.ID)
		return nil
	},
}

var planListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all plans",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := workflow.NewPlanStore(cfg.PlansDir())
		if err != nil {
			return err
		}

		plans, err := store.List()
		if err != nil {
			return err
		}

		if len(plans) == 0 {
			fmt.Println("No plans found. Create one with: agentmux plan create <title>")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tCREATED")
		fmt.Fprintln(w, "--\t-----\t------\t-------")

		for _, p := range plans {
			title := p.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				p.ID, title,
				workflow.FormatStatus(p.Status),
				p.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		w.Flush()
		return nil
	},
}

var planShowCmd = &cobra.Command{
	Use:   "show <plan-id>",
	Short: "Show plan details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := workflow.NewPlanStore(cfg.PlansDir())
		if err != nil {
			return err
		}

		plan, err := store.Get(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Plan: %s\n", plan.ID)
		fmt.Printf("Title: %s\n", plan.Title)
		fmt.Printf("Status: %s\n", workflow.FormatStatus(plan.Status))
		fmt.Printf("Created: %s\n", plan.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", plan.UpdatedAt.Format("2006-01-02 15:04:05"))

		if plan.Description != "" {
			fmt.Printf("\nDescription:\n  %s\n", plan.Description)
		}
		if plan.Agent != "" {
			fmt.Printf("Agent: %s\n", plan.Agent)
		}
		if plan.RejectReason != "" {
			fmt.Printf("Reject Reason: %s\n", plan.RejectReason)
		}

		return nil
	},
}

var planApproveCmd = &cobra.Command{
	Use:   "approve <plan-id>",
	Short: "Approve a draft plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := workflow.NewPlanStore(cfg.PlansDir())
		if err != nil {
			return err
		}

		if err := store.Approve(args[0]); err != nil {
			return err
		}

		fmt.Printf("✓ Plan %s approved\n", args[0])
		return nil
	},
}

var planRejectCmd = &cobra.Command{
	Use:   "reject <plan-id>",
	Short: "Reject a draft plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reason, _ := cmd.Flags().GetString("reason")

		store, err := workflow.NewPlanStore(cfg.PlansDir())
		if err != nil {
			return err
		}

		if err := store.Reject(args[0], reason); err != nil {
			return err
		}

		fmt.Printf("✗ Plan %s rejected\n", args[0])
		if reason != "" {
			fmt.Printf("  Reason: %s\n", reason)
		}
		return nil
	},
}

var planDeleteCmd = &cobra.Command{
	Use:   "delete <plan-id>",
	Short: "Delete a plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := workflow.NewPlanStore(cfg.PlansDir())
		if err != nil {
			return err
		}

		if err := store.Delete(args[0]); err != nil {
			return err
		}

		fmt.Printf("✓ Plan %s deleted\n", args[0])
		return nil
	},
}

func init() {
	planCreateCmd.Flags().StringP("description", "d", "", "plan description")
	planRejectCmd.Flags().StringP("reason", "r", "", "reason for rejection")

	planCmd.AddCommand(planCreateCmd)
	planCmd.AddCommand(planListCmd)
	planCmd.AddCommand(planShowCmd)
	planCmd.AddCommand(planApproveCmd)
	planCmd.AddCommand(planRejectCmd)
	planCmd.AddCommand(planDeleteCmd)

	rootCmd.AddCommand(planCmd)
}
