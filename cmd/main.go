package main

import (
	"fmt"
	"log"
	"mini-hpc-manager/pkg/job"
	"mini-hpc-manager/pkg/scheduler"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func main() {

	// CLI
	var rootCmd = &cobra.Command{
		Use:   "mini-hpc",
		Short: "Mini HPC Manager",
		Long:  `A simple job manager for HPC and cloud environments using Go and Docker.`,
	}

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new job",
	Long:  `Add a new job to the scheduler.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Usage: mini-hpc add <docker-image> <command>")
			return
		}

		image := args[0]
		command := args[1]

		job := job.Job{
			ID:      uuid.New().String(),
			Name:    "Job-" + uuid.New().String(),
			Command: command,
			Status:  job.JobStatusPending,
			CPU:     1,                  // Defaults, can be customized later
			Memory:  1024 * 1024 * 1024, // Defaults, can be customized later
			Image:   image,
			Log:     "",
		}

		// Add the job to the scheduler
		scheduler := scheduler.NewScheduler()
		scheduler.AddJob(job)
		fmt.Printf("[scheduler] -- Job added: %s\n", job.ID)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all jobs",
	Run: func(cmd *cobra.Command, args []string) {
		scheduler := scheduler.NewScheduler()

		if len(scheduler.Queue) == 0 {
			fmt.Println("[scheduler] -- No jobs in the queue")
			return
		}

		for i, j := range scheduler.Queue {
			fmt.Printf("[scheduler] -- [%d] ID: %s, Image: %s, Command: %s, Status: %s\n", i+1, j.ID, j.Image, j.Command, j.Status)
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the next job in the queue",
	Run: func(cmd *cobra.Command, args []string) {
		scheduler := scheduler.NewScheduler()

		if len(scheduler.Queue) == 0 {
			fmt.Println("[scheduler] -- No jobs in the queue")
			return
		}

		scheduler.Run()
	},
}
