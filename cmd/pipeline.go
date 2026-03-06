package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cqi/my_agentmux/internal/agent"
	"github.com/cqi/my_agentmux/internal/config"
	"github.com/cqi/my_agentmux/internal/session"
	"github.com/spf13/cobra"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Manage and run agent pipelines",
	Long:  `Run a predefined sequence of agents from your project configuration.`,
}

var pipelineRunCmd = &cobra.Command{
	Use:   "run <pipeline-name>",
	Short: "Run a pipeline sequence",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pipelineName := args[0]

		workDir, err := os.Getwd()
		if err != nil {
			return err
		}

		projectCfg, err := config.LoadProjectConfig(workDir)
		if err != nil {
			return err
		}
		if projectCfg == nil {
			return fmt.Errorf("no project configuration found. Run 'agentmux init' in your project root")
		}

		pipeline, exists := projectCfg.Pipelines[pipelineName]
		if !exists || len(pipeline) == 0 {
			return fmt.Errorf("pipeline %q not found or empty in project config", pipelineName)
		}

		activeCfg := config.MergeProjectConfig(cfg, projectCfg)

		ctx := context.Background()
		mgr, err := session.NewManager(activeCfg)
		if err != nil {
			return err
		}
		runner := agent.NewRunner(activeCfg, mgr)

		fmt.Printf("▶ Starting pipeline %q: %v\n", pipelineName, pipeline)

		for i, agentName := range pipeline {
			fmt.Printf("\n[%d/%d] Starting agent: %s\n", i+1, len(pipeline), agentName)

			// Start the agent
			opts := agent.LaunchOptions{
				Name:    agentName, // Use the definition name as the session name
				WorkDir: projectCfg.DefaultWorkDir,
			}

			// If the agent doesn't exist as a definition, Launch will use DefaultAgentType
			_, err := runner.Launch(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to launch agent %q: %w", agentName, err)
			}

			fmt.Printf("✓ Agent %q running. Waiting for completion...\n", agentName)

			// Wait for agent to finish
			for {
				sess, err := mgr.Get(ctx, agentName)
				if err != nil {
					// Session no longer exists
					break
				}
				if sess.Status != "running" {
					// Clean up state
					_ = mgr.Destroy(ctx, agentName)
					break
				}
				time.Sleep(1 * time.Second)
			}

			fmt.Printf("✓ Agent %q finished.\n", agentName)
		}

		fmt.Printf("\n🎉 Pipeline %q completed successfully!\n", pipelineName)
		return nil
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineRunCmd)
	rootCmd.AddCommand(pipelineCmd)
}
