package scheduler

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
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

	ctx := context.Background()

	// Pull the Docker image
	reader, err := s.dockerClient.ImagePull(ctx, nextJob.Image, image.PullOptions{})
	if err != nil {
		log.Println("[scheduler] -- Error pulling image: ", err)
		nextJob.Status = job.JobStatusFailed
		return
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)

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
		return
	}

	// Start the container
	if err := s.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Println("[scheduler] -- Error starting container: ", err)
		nextJob.Status = job.JobStatusFailed
		return
	}

	// Wait for the container to finish
	statusCh, errCh := s.dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Println("[scheduler] -- Error waiting for container: ", err)
			nextJob.Status = job.JobStatusFailed
			return
		}
	case <-statusCh:
	}

	// Capture the container's logs
	out, err := s.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Println("[scheduler] -- Error getting container logs: ", err)
		nextJob.Status = job.JobStatusFailed
		return
	}

	// Read all logs from the container
	logOutput, err := ioutil.ReadAll(out)
	if err != nil {
		log.Println("[scheduler] -- Error reading container logs: ", err)
		nextJob.Status = job.JobStatusFailed
		return
	}

	// Update the job status and log
	nextJob.Log = string(logOutput)
	nextJob.Status = job.JobStatusComplete
	fmt.Printf("[scheduler] -- Job %s complete\n", nextJob.ID)
	fmt.Printf("[scheduler] -- Job log:\n%s\n", nextJob.Log)

	// Clean up the container
	if err := s.dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
		log.Println("[scheduler] -- Error removing container: ", err)
		nextJob.Status = job.JobStatusFailed
		return
	}
}
