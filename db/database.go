package db

import (
	"database/sql"
	"mini-hpc-manager/pkg/job"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath string = "mini-hpc.db"

var db *sql.DB

// InitDatabase initializes the SQLite database
func InitDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Create the jobs table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		name TEXT,
		command TEXT,
		status TEXT,
		cpu INTEGER,
		memory INTEGER,
		image TEXT,
		log TEXT
	)`)
	if err != nil {
		return err
	}

	return nil
}

// AddJob adds a new job to the database
func AddJob(j job.Job) error {
	_, err := db.Exec("INSERT INTO jobs (id, name, command, status, cpu, memory, image, log) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		j.ID, j.Name, j.Command, j.Status, j.CPU, j.Memory, j.Image, j.Log)
	return err
}

// UpdateJob updates an existing job in the database
func UpdateJob(j job.Job) error {
	_, err := db.Exec("UPDATE jobs SET name = ?, command = ?, status = ?, cpu = ?, memory = ?, image = ?, log = ? WHERE id = ?",
		j.Name, j.Command, j.Status, j.CPU, j.Memory, j.Image, j.Log, j.ID)
	return err
}

// LoadQueue loads all jobs from the database
func LoadQueue() ([]job.Job, error) {
	rows, err := db.Query("SELECT * FROM jobs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []job.Job
	for rows.Next() {
		var j job.Job
		err := rows.Scan(&j.ID, &j.Name, &j.Command, &j.Status, &j.CPU, &j.Memory, &j.Image, &j.Log)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}

// CloseDatabase closes the SQLite database
func CloseDatabase() error {
	return db.Close()
}
