package scheduler

import (
	"context"
	"fmt"
	"io"
	"log"
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
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &Scheduler{
		Queue:        []job.Job{},
		dockerClient: cli,
	}
}

func (s *Scheduler) AddJob(j job.Job) {
	s.Queue = append(s.Queue, j)
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

	context := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("[scheduler] -- Error creating Docker client: ", err)
		nextJob.Status = job.JobStatusFailed
		panic(err)
	}

	defer cli.Close()

	// Pull the Docker image
	// cli.ImagePull is asynchronous, so we need to wait for the pull to complete

	reader, err := cli.ImagePull(context, nextJob.Image, image.PullOptions{})
	if err != nil {
		log.Println("[scheduler] -- Error pulling image: ", err)
		nextJob.Status = job.JobStatusFailed
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	// Create the container from the image
	// We need to set the container's resources (CPU and memory) based on the job's requirements
	// We also need to set the container's command to the job's command

	resp, err := cli.ContainerCreate(context, &container.Config{
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
		panic(err)
	}

	// Start the container
	if err := cli.ContainerStart(context, resp.ID, container.StartOptions{}); err != nil {
		log.Println("[scheduler] -- Error starting container: ", err)
		nextJob.Status = job.JobStatusFailed
		panic(err)
	}

	// Wait for the container to finish
	statusCh, errCh := cli.ContainerWait(context, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Println("[scheduler] -- Error waiting for container: ", err)
			nextJob.Status = job.JobStatusFailed
			panic(err)
		}
	case <-statusCh:
	}

	// Capture the container's logs
	out, err := cli.ContainerLogs(context, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Println("[scheduler] -- Error getting container logs: ", err)
		nextJob.Status = job.JobStatusFailed
		panic(err)
	}

	// Update the job's log
	var logOutput []byte
	if _, err := out.Read(logOutput); err != nil {
		log.Println("[scheduler] -- Error reading container logs: ", err)
		nextJob.Status = job.JobStatusFailed
		panic(err)
	}

	// Update the job status and log
	nextJob.Log = string(logOutput)
	nextJob.Status = job.JobStatusComplete
	fmt.Printf("[scheduler] -- Job %s complete\n", nextJob.ID)
	fmt.Printf("[scheduler] -- Job log:\n%s\n", nextJob.Log)

	// Recursively run the next job
	s.Run()

	// Clean up the container
	if err := cli.ContainerRemove(context, resp.ID, container.RemoveOptions{}); err != nil {
		log.Println("[scheduler] -- Error removing container: ", err)
		nextJob.Status = job.JobStatusFailed
		panic(err)
	}
}
