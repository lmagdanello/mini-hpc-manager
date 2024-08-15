# mini-hpc-manager

> Welcome to **Mini HPC Manager!** 🎉

Mini HPC Manager is a simple tool designed to manage and execute jobs in HPC (High Performance Computing) and Cloud environments using Go and Docker. It provides a CLI for adding, listing, and running jobs with Docker containers.

## features

- Job Scheduling: Add jobs to a queue and specify their Docker images and commands.
- Database Integration: Stores job details and status using SQLite.
- Docker Integration: Handles job execution via Docker containers.
- Status Tracking: Updates and tracks job status (Pending, Running, Complete, Failed).
- Log Collection: Captures and stores job logs for review.

## getting started

1. Clone the Repository:

```bash
git clone https://github.com/yourusername/mini-hpc-manager.git
cd mini-hpc-manager
```

2. Install Dependencies:

Make sure you have Go installed, and then install the required dependencies:

```bash
go mod download
```

3. Build the Project:

```bash
go build -o mini-hpc
```

4. Add a Job:

```bash
./mini-hpc add <docker-image> <command>
```

5. List Jobs:

```bash
./mini-hpc list
```
6. Run a Job:

```bash
./mini-hpc run <job-id>
```

## architecture

```lua
+---------------------+       +----------------------+
|                     |       |                      |
|    CLI Commands     |       |    Scheduler         |
|    (add, list, run) |       |                      |
|                     |       |                      |
+----------+----------+       +----------+-----------+
           |                           |
           |                           |
           v                           v
+----------+----------+       +----------+-----------+
|                     |       |                      |
|  SQLite Database    |       |   Docker Engine      |
|                     |       |                      |
+---------------------+       +----------------------+
           ^
           |
           |
           v
+----------+----------+
|                     |
|    Job Data Model   |
|                     |
+---------------------+
```

- Release binary!
- CLI Commands: The command-line interface allows you to add, list, and run jobs.
- Scheduler: Manages job scheduling, execution, and status updates.
- SQLite Database: Stores job details and status.
- Docker Engine: Executes the jobs in Docker containers.
- Job Data Model: Represents the job structure used by both the scheduler and database.

## future features?

- Job Priority: Implement job prioritization for more advanced scheduling.
- Job Dependencies: Support job dependencies and chained execution.
- User Authentication: Add authentication for user management.
- Web Interface: Build a web-based UI for easier job management.
- Enhanced Logging: Improve log handling and storage.

## License
This project is licensed under the MIT License. See the LICENSE file for details.

