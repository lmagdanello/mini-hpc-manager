package main

import (
	"fmt"
	"log"
	"mini-hpc-manager/db"
	"mini-hpc-manager/pkg/job"
	"mini-hpc-manager/pkg/scheduler"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func main() {
	// Initialize the database
	if err := db.InitDatabase(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.CloseDatabase()

	// CLI setup
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
		if len(args) < 2 {
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
		s := scheduler.NewScheduler()
		defer db.CloseDatabase() // Ensure the database is closed when done
		s.AddJob(job)
		fmt.Printf("[scheduler] -- Job added: %s\n", job.ID)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all jobs",
	Run: func(cmd *cobra.Command, args []string) {
		s := scheduler.NewScheduler()
		defer db.CloseDatabase() // Ensure the database is closed when done

		if len(s.Queue) == 0 {
			fmt.Println("[scheduler] -- No jobs in the queue")
			return
		}

		for i, j := range s.Queue {
			fmt.Printf("[scheduler] -- [%d] ID: %s, Image: %s, Command: %s, Status: %s\n", i+1, j.ID, j.Image, j.Command, j.Status)
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the next job in the queue",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jobID := args[0] // Get the job ID from the command argument

		s := scheduler.NewScheduler()
		defer db.CloseDatabase() // Ensure the database is closed when done

		// Run the specific job
		s.Run(jobID)
	},
}
