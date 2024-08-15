package scheduler

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mini-hpc-manager/db"
	"mini-hpc-manager/pkg/job"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type Scheduler struct {
	Queue        []job.Job
	dockerClient *client.Client
}

func NewScheduler() *Scheduler {

	// Initialize the Database
	if err := db.InitDatabase(); err != nil {
		log.Fatalf("[scheduler] -- Error initializing database: %v", err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Load the queue from the database
	jobs, err := db.LoadQueue()
	if err != nil {
		log.Fatalf("[scheduler] -- Error loading queue: %v", err)
	}

	return &Scheduler{
		Queue:        jobs,
		dockerClient: cli,
	}
}

func (s *Scheduler) AddJob(j job.Job) {
	s.Queue = append(s.Queue, j)
	if err := db.AddJob(j); err != nil {
		log.Printf("[scheduler] -- Error adding job to database: %v", err)
	}
}

func (s *Scheduler) Run() {
	if len(s.Queue) == 0 {
		fmt.Println("[scheduler] -- No jobs to run")
		return
	}

	// Get the next job in the queue
	nextJob := s.Queue[0]
	s.Queue = s.Queue[1:]

	fmt.Printf("[scheduler] -- Running job: %s\n", nextJob.ID)
	fmt.Printf("[scheduler] -- Container image: %s\n", nextJob.Image)

	// Update job status
	nextJob.Status = job.JobStatusRunning
	if err := db.UpdateJob(nextJob); err != nil {
		log.Printf("[scheduler] -- Error updating job status in database: %v", err)
	}

	ctx := context.Background()

	// Pull the Docker image
	reader, err := s.dockerClient.ImagePull(ctx, nextJob.Image, image.PullOptions{})
	if err != nil {
		log.Println("[scheduler] -- Error pulling image: ", err)
		nextJob.Status = job.JobStatusFailed
		return
	} else {
		// Ensure the image is pulled before continuing
		// This is necessary because the image pull is done asynchronously
		// and the container creation will fail if the image is not pulled
		// before the container is created
		defer reader.Close()
		io.Copy(os.Stdout, reader)
	}

	// Create the container
	resp, err := s.dockerClient.ContainerCreate(ctx, &container.Config{
		Image: nextJob.Image,
		Cmd:   []string{nextJob.Command},
		Tty:   false,
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory:   int64(nextJob.Memory),
			NanoCPUs: int64(nextJob.CPU * 1e9),
		},
	}, nil, nil, "")
	if err != nil {
		log.Println("[scheduler] -- Error creating container: ", err)
		nextJob.Status = job.JobStatusFailed
		db.UpdateJob(nextJob)
		return
	}

	// Start the container
	if err := s.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Println("[scheduler] -- Error starting container: ", err)
		nextJob.Status = job.JobStatusFailed
		db.UpdateJob(nextJob)
		return
	}

	// Wait for the container to finish
	statusCh, errCh := s.dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Println("[scheduler] -- Error waiting for container: ", err)
			nextJob.Status = job.JobStatusFailed
			db.UpdateJob(nextJob)
			return
		}
	case <-statusCh:
	}

	// Capture the container's logs
	out, err := s.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Println("[scheduler] -- Error getting container logs: ", err)
		nextJob.Status = job.JobStatusFailed
		db.UpdateJob(nextJob)
		return
	}

	// Read all logs from the container
	logOutput, err := ioutil.ReadAll(out)
	if err != nil {
		log.Println("[scheduler] -- Error reading container logs: ", err)
		nextJob.Status = job.JobStatusFailed
		db.UpdateJob(nextJob)
		return
	}

	// Update the job status and log
	nextJob.Log = string(logOutput)
	nextJob.Status = job.JobStatusComplete
	fmt.Printf("[scheduler] -- Job %s complete\n", nextJob.ID)
	fmt.Printf("[scheduler] -- Job log:\n%s\n", nextJob.Log)

	// Update the job in the database
	if err := db.UpdateJob(nextJob); err != nil {
		log.Printf("[scheduler] -- Error updating job in database: %v", err)
	}

	// Clean up the container
	if err := s.dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
		log.Println("[scheduler] -- Error removing container: ", err)
		nextJob.Status = job.JobStatusFailed
		return
	}
}

func CloseScheduler() {
	if err := db.CloseDatabase(); err != nil {
		log.Fatalf("[scheduler] -- Error closing database: %v", err)
	}
}
