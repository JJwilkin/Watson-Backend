package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Job represents a task to be processed
type Job struct {
	ID              string          `json:"id"`
	BatchID         string          `json:"batch_id,omitempty"`
	Type            string          `json:"type"`
	Data            json.RawMessage `json:"data"`
	OnCompleteJobId string          `json:"on_complete_job_id,omitempty"` // function name to call when job is complete. we will map this in a switch statement
	CreatedAt       time.Time       `json:"created_at"`
	StartedAt       *time.Time      `json:"started_at,omitempty"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	Progress        string          `json:"progress"` // pending, processing, completed, failed
	Error           string          `json:"error,omitempty"`
}

// EnqueueRequest represents the request body for enqueueing jobs
type EnqueueRequest struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// EnqueueResponse represents the response when enqueueing a job
type EnqueueResponse struct {
	Success bool   `json:"success"`
	JobID   string `json:"job_id,omitempty"`
	Message string `json:"message,omitempty"`
}

// JobProcessor handles job processing
type JobProcessor struct {
	rdb        *redis.Client
	httpClient *http.Client
}

var bulkJobQueue = "bulk_job_queue"
var pendingJobQueue = "pending_job_queue"
var processingJobQueue = "processing_job_queue"
var completedJobQueue = "completed_job_queue"
var failedJobQueue = "failed_job_queue"

// NewJobProcessor creates a new job processor
func NewJobProcessor(addr string) *JobProcessor {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	// Load client certificates
	cert, err := tls.LoadX509KeyPair("./certs/certificate.pem", "./certs/private_key.pem")
	if err != nil {
		log.Fatal("Failed to load client certificates:", err)
	}

	// Create TLS config with client certificates
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Create HTTP client with custom transport
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	return &JobProcessor{rdb: rdb, httpClient: httpClient}
}

func (jp *JobProcessor) CreateBulkJob(jobs []*Job, conCompleteJobId string) (*Job, error) {
	jobIds := []string{}
	for _, job := range jobs {
		err := jp.StoreJob(job)
		if err != nil {
			return nil, fmt.Errorf("failed to store job: %w", err)
		}
		jobIds = append(jobIds, job.ID)
		jp.EnqueueJobId(job.ID)
	}

	jobIdsJSON, err := json.Marshal(jobIds)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job IDs: %w", err)
	}

	bulkJob := &Job{
		ID:              uuid.New().String(),
		Type:            "bulk_job",
		Data:            jobIdsJSON,
		CreatedAt:       time.Now(),
		StartedAt:       nil,
		CompletedAt:     nil,
		Progress:        "pending",
		OnCompleteJobId: conCompleteJobId,
	}
	err = jp.StoreJob(bulkJob)
	if err != nil {
		return nil, fmt.Errorf("failed to store bulk job: %w", err)
	}

	return bulkJob, nil
}

func (jp *JobProcessor) CreateJob(jobType string, data json.RawMessage) (*Job, error) {
	job := Job{
		ID:          uuid.New().String(),
		Type:        jobType,
		Data:        data,
		CreatedAt:   time.Now(),
		StartedAt:   nil,
		CompletedAt: nil,
		Progress:    "pending",
	}
	return &job, nil
}

func (jp *JobProcessor) CreateAndStoreJob(jobType string, data json.RawMessage) (*Job, error) {
	job, err := jp.CreateJob(jobType, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}
	err = jp.StoreJob(job)
	if err != nil {
		return nil, fmt.Errorf("failed to store job: %w", err)
	}
	return job, nil
}

func (jp *JobProcessor) StoreJob(job *Job) error {
	ctx := context.Background()
	jobKey := job.ID
	err := jp.rdb.HSet(ctx, jobKey, map[string]interface{}{
		"id":                 job.ID,
		"batch_id":           job.BatchID,
		"type":               job.Type,
		"data":               string(job.Data),
		"on_complete_job_id": job.OnCompleteJobId,
		"created_at":         job.CreatedAt.Format(time.RFC3339),
		"progress":           job.Progress,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to store job details: %w", err)
	}
	return nil
}

func (jp *JobProcessor) GetJob(jobID string) (*Job, error) {
	ctx := context.Background()
	result, err := jp.rdb.HGetAll(ctx, jobID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job details from redis: %w", err)
	}

	job := &Job{ID: jobID}
	for key, value := range result {
		switch key {
		case "type":
			job.Type = value
		case "batch_id":
			job.BatchID = value
		case "data":
			job.Data = json.RawMessage(value)
		case "progress":
			job.Progress = value
		case "on_complete_job_id":
			job.OnCompleteJobId = value
		case "created_at":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				job.CreatedAt = t
			}
		case "started_at":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				job.StartedAt = &t
			}
		case "completed_at":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				job.CompletedAt = &t
			}
		}
	}

	return job, nil
}

func (jp *JobProcessor) UpdateJobStatus(jobId string, status string) error {
	ctx := context.Background()
	err := jp.rdb.HSet(ctx, jobId, "progress", status).Err()
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}

// This assumes the jobs have been stored in redis
func (jp *JobProcessor) EnqueueBulkJob(bulkJobId string) error {
	ctx := context.Background()

	err := jp.rdb.LPush(ctx, bulkJobQueue, bulkJobId).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue bulk job: %w", err)
	}

	return nil
}

func (jp *JobProcessor) CheckBulkJobStatus(bulkJobId string) (int, error) {
	// ctx := context.Background()

	bulkJob, err := jp.GetJob(bulkJobId)
	if err != nil {
		return -1, fmt.Errorf("failed to get bulk job: %w", err)
	}
	jobIds := []string{}
	err = json.Unmarshal(bulkJob.Data, &jobIds)
	if err != nil {
		return -1, fmt.Errorf("failed to unmarshal job IDs: %w", err)
	}

	count := 0
	for _, jobId := range jobIds {
		job, err := jp.GetJob(jobId)
		if err != nil {
			return -1, fmt.Errorf("failed to get job: %w", err)
		}

		if job.Progress == "pending" || job.Progress == "processing" {
			count++
		}
	}

	// if count == 0 {
	// 	jp.rdb.LRem(ctx, bulkJobQueue, 1, bulkJob.ID).Err()
	// }

	return count, nil
}

// This assumes the job has been stored in redis
func (jp *JobProcessor) EnqueueJobId(jobId string) error {
	ctx := context.Background()

	// Add job to the queue (using LPUSH to add to the left of the list)
	err := jp.rdb.LPush(ctx, pendingJobQueue, jobId).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}
	return nil
}

func (jp *JobProcessor) EnqueueJob(job *Job) error {

	// Store job details in Redis hash for status tracking
	err := jp.StoreJob(job)
	if err != nil {
		return fmt.Errorf("failed to store job: %w", err)
	}

	jp.EnqueueJobId(job.ID)

	log.Printf("✅ Enqueued job: %s (Type: %s)", job.ID, job.Type)
	return nil
}

func (jp *JobProcessor) DequeueJob() (*Job, error) {
	ctx := context.Background()

	// moves job from pending to processing
	jobId, err := jp.rdb.BLMove(ctx, pendingJobQueue, processingJobQueue, "LEFT", "RIGHT", 5*time.Second).Result()
	if err != nil {
		if err == redis.Nil {
			// No jobs available, this is normal
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	job, err := jp.GetJob(jobId)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	jp.UpdateJobStatus(jobId, "processing")

	return job, nil
}

func (jp *JobProcessor) DequeueOrRequeueBulkJob() (bool, error) {
	ctx := context.Background()

	bulkJobId, err := jp.rdb.RPop(ctx, bulkJobQueue).Result()
	if err != nil {
		if err == redis.Nil {
			// No jobs left in the queue
			return false, nil
		}
		return false, fmt.Errorf("failed to dequeue bulk job: %w", err)
	}
	count, err := jp.CheckBulkJobStatus(bulkJobId)
	if err != nil {
		return false, fmt.Errorf("failed to check bulk job status: %w", err)
	}
	if count != 0 {
		jp.EnqueueBulkJob(bulkJobId)
		return false, nil
	}

	bulkJob, _ := jp.GetJob(bulkJobId)
	if bulkJob.OnCompleteJobId != "" {
		jp.EnqueueJobId(bulkJob.OnCompleteJobId)
	}
	return true, nil
}

func (jp *JobProcessor) MarkJobAsCompleted(jobId string) error {
	ctx := context.Background()

	// Remove the job from the processing queue
	if _, err := jp.rdb.LRem(ctx, processingJobQueue, 1, jobId).Result(); err != nil {
		return fmt.Errorf("failed to remove job from processing queue: %w", err)
	}
	// Add the job to the completed queue
	if err := jp.rdb.LPush(ctx, completedJobQueue, jobId).Err(); err != nil {
		return fmt.Errorf("failed to add job to completed queue: %w", err)
	}
	jp.UpdateJobStatus(jobId, "completed")
	return nil
}

func (jp *JobProcessor) MarkJobAsFailed(jobId string) error {
	ctx := context.Background()

	if _, err := jp.rdb.LRem(ctx, processingJobQueue, 1, jobId).Result(); err != nil {
		return fmt.Errorf("failed to remove from processing: %w", err)
	}
	jp.UpdateJobStatus(jobId, "failed")
	return jp.rdb.LPush(ctx, failedJobQueue, jobId).Err()
}
