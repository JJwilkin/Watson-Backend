package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"watson/database"
	"watson/plaid"
)

var ctx = context.Background()

// MarkJobStatus updates the status of a job

// UpdateJobProgress updates the progress of a job
// func (jp *JobProcessor) UpdateJobProgress(jobID string, progress string) error {
// 	return jp.rdb.HSet(ctx, jobID, "progress", progress).Err()
// }

// GetJobStatus retrieves the status of a job
func (jp *JobProcessor) GetJobStatus(jobID string) (string, error) {
	result, err := jp.rdb.HGet(ctx, jobID, "status").Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (jp *JobProcessor) GetJobStatusByID(jobId string) (string, error) {
	result, err := jp.rdb.HGet(ctx, jobId, "status").Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

// GetJob retrieves a complete job by ID
// func (jp *JobProcessor) GetJob(jobID string) (*Job, error) {
// 	jobKey := fmt.Sprintf("job:%s", jobID)
// 	result, err := jp.rdb.HGetAll(ctx, jobKey).Result()
// 	if err != nil {
// 		return nil, err
// 	}

// 	job := &Job{ID: jobID}
// 	for key, value := range result {
// 		switch key {
// 		case "type":
// 			job.Type = value
// 		case "progress":
// 			if progress, err := strconv.Atoi(value); err == nil {
// 				job.Progress = progress
// 			}
// 		case "error":
// 			job.Error = value
// 		case "created_at":
// 			if t, err := time.Parse(time.RFC3339, value); err == nil {
// 				job.CreatedAt = t
// 			}
// 		case "started_at":
// 			if t, err := time.Parse(time.RFC3339, value); err == nil {
// 				job.StartedAt = &t
// 			}
// 		case "completed_at":
// 			if t, err := time.Parse(time.RFC3339, value); err == nil {
// 				job.CompletedAt = &t
// 			}
// 		}
// 	}

// 	return job, nil
// }

func (jp *JobProcessor) functionMap(name string) func(job *Job) error {
	switch name {
	case "job_monitor":
		return jp.processJobMonitor
	case "process_daily_balance":
		return jp.processDailyBalance
	case "fetch_plaid_transactions":
		return jp.processFetchPlaidTransactions
	case "fetch_all_new_transactions":
		return jp.fetchAllNewTransactions
	case "initial_plaid_sync":
		return jp.processInitialPlaidSync
	}
	return nil
}

// func (jp *JobProcessor) createAndEnqueueJobBatch(jobs []Job, onComplete string, onCompleteData json.RawMessage) error {
// 	// for each job, enqueue it, adding a batch_id to the jobs, and then create a job monitor job with the batch_id
// 	// this will check the status of the jobs, and re-enqueue the monitor job if any of the jobs are not completed
// 	// after all jobs are completed, takes in a callback function to run after all jobs are completed
// 	batchID := fmt.Sprintf("batch_%d", time.Now().UnixNano())
// 	batchKey := fmt.Sprintf("batch:%s", batchID)
// 	for i := range jobs {
// 		jobs[i].BatchID = batchID
// 		jobPtr := &jobs[i]
// 		jp.EnqueueJob(jobPtr)
// 		jp.rdb.LPush(ctx, batchKey, jobs[i].ID)
// 	}

// 	err := jp.createAndEnqueueJob("job_monitor", onCompleteData, batchID, onComplete)
// 	if err != nil {
// 		return fmt.Errorf("failed to create and enqueue job monitor: %w", err)
// 	}
// 	return nil
// }

// EnqueueJob adds a job to the queue
func (jp *JobProcessor) createAndEnqueueJob(jobType string, data json.RawMessage, batchID string, onCompleteJobId string) error {
	job := &Job{
		ID:              fmt.Sprintf("job_%d", time.Now().UnixNano()),
		BatchID:         batchID,
		Type:            jobType,
		Data:            data,
		CreatedAt:       time.Now(),
		Progress:        "pending",
		OnCompleteJobId: onCompleteJobId,
	}

	return jp.EnqueueJob(job)
}

// Note: DequeueJob is implemented in jobqueue.go

// ProcessJob handles the actual job processing
func (jp *JobProcessor) ProcessJob(job *Job) error {
	log.Printf("🔄 Processing job: %s (Type: %s)", job.ID, job.Type)

	// Update progress to 10% (started)
	jp.UpdateJobStatus(job.ID, "processing")

	var err error
	switch job.Type {
	case "hello_world":
		err = jp.processHelloWorld(job)
	case "print_message":
		err = jp.processPrintMessage(job)
	case "initial_plaid_sync":
		err = jp.processInitialPlaidSync(job)
	case "fetch_plaid_transactions":
		err = jp.processFetchPlaidTransactions(job)
	case "fetch_all_new_transactions":
		err = jp.fetchAllNewTransactions(job)
	case "process_daily_balance":
		err = jp.processDailyBalance(job)
	case "job_monitor":
		err = jp.processJobMonitor(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	// Update progress to 90% (almost done)

	// Mark job as completed or failed
	if err != nil {
		// jp.MarkJobStatus(job.ID, "failed")
		jp.MarkJobAsFailed(job.ID)
		jp.rdb.HSet(ctx, fmt.Sprintf("job:%s", job.ID), "error", err.Error())
		return err
	}

	// Update progress to 100% (completed)
	jp.MarkJobAsCompleted(job.ID)

	// enqueue the on complete job if it exists
	if job.OnCompleteJobId != "" {
		jp.EnqueueJobId(job.OnCompleteJobId)
		log.Printf("✅ Enqueued on complete job: %s", job.OnCompleteJobId)
	}

	return nil
}

// processHelloWorld handles hello world jobs
func (jp *JobProcessor) processHelloWorld(job *Job) error {
	fmt.Printf("🌍 Hello World! Job ID: %s, Data: %s\n", job.ID, string(job.Data))

	// Simulate some processing time
	time.Sleep(1 * time.Second)

	log.Printf("✅ Completed hello world job: %s", job.ID)
	return nil
}

// processPrintMessage handles print message jobs
func (jp *JobProcessor) processPrintMessage(job *Job) error {
	fmt.Printf("📝 Message: %s (Job ID: %s)\n", string(job.Data), job.ID)

	// Simulate some processing time
	time.Sleep(50 * time.Millisecond)

	log.Printf("✅ Completed print message job: %s", job.ID)
	return nil
}

// PLAID

func (jp *JobProcessor) processInitialPlaidSync(job *Job) error {
	log.Printf("🔄 Processing initial Plaid sync job: %s", job.ID)

	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}
	// log.Printf("🔄 Job data: %v", jobData)
	accessToken := jobData["access_token"].(string)
	if accessToken == "" {
		return fmt.Errorf("access token is empty")
	}
	isSandbox := strings.HasPrefix(accessToken, "access-sandbox")
	log.Printf("access token: %s", accessToken)
	log.Printf("is this sandbox? %v", isSandbox)
	accounts, err := plaid.GetAccounts(accessToken, isSandbox)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	log.Printf("✅ Fetched %d accounts from Plaid", len(accounts))
	plaidTokenID, userID, err := database.GetUserIdFromAccessToken(accessToken)
	if err != nil {
		return fmt.Errorf("failed to get user id from access token: %w", err)
	}

	err = database.MarkPlaidTokenAsProcessed(plaidTokenID)
	if err != nil {
		return fmt.Errorf("failed to mark plaid token as processed: %w", err)
	}

	err = database.CreatePlaidAccount(userID, plaidTokenID, accounts)
	if err != nil {
		return fmt.Errorf("failed to create plaid account: %w", err)
	}

	log.Printf("✅ Completed initial Plaid sync job: %s", job.ID)

	// enqueue job to fetch transactions for each plaid account
	newJobData := map[string]interface{}{
		"user_id":    userID,
		"month_year": GetCurrentMonthYear(),
	}
	newJobDataJSON, _ := json.Marshal(newJobData)
	newJob, _ := jp.CreateJob("fetch_all_new_transactions", newJobDataJSON)
	jp.EnqueueJob(newJob)
	return nil
}

func (jp *JobProcessor) fetchAllNewTransactions(job *Job) error {
	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}
	// log.Printf("🔄 Job data: %v", jobData)
	userID := int(jobData["user_id"].(float64))
	monthYear := int(jobData["month_year"].(float64))
	// get plaid account by UserId
	accounts, err := database.GetPlaidAccountsByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get plaid accounts by user id: %w", err)
	}
	jobs := []*Job{}
	for _, accountID := range accounts {
		jobData := map[string]interface{}{
			"account_id": accountID,
			"user_id":    userID,
			"month_year": monthYear,
		}
		jobDataJSON, _ := json.Marshal(jobData)
		jobs = append(jobs, &Job{
			Type: "fetch_all_new_transactions",
			Data: jobDataJSON,
		})
	}
	// jp.createAndEnqueueJobBatch(jobs, "fetch_all_new_transactions", json.RawMessage(`{}`))
	//callback job
	callBackJobData := map[string]interface{}{
		"user_id":    userID,
		"month_year": monthYear,
	}
	callBackJobDataJSON, _ := json.Marshal(callBackJobData)
	callbackJob, _ := jp.CreateAndStoreJob("process_daily_balance", callBackJobDataJSON)
	bulkJob, err := jp.CreateBulkJob(jobs, callbackJob.ID)
	if err != nil {
		return fmt.Errorf("failed to create bulk job: %w", err)
	}
	jp.EnqueueBulkJob(bulkJob.ID)
	return nil
}

func (jp *JobProcessor) checkBulkJobStatus(job *Job) error {
	bulkJobId := job.ID
	onCompleteJobId := job.OnCompleteJobId
	count, err := jp.CheckBulkJobStatus(bulkJobId)
	if err != nil {
		return fmt.Errorf("failed to check bulk job status: %w", err)
	}

	if count == 0 {
		jp.rdb.LRem(ctx, bulkJobQueue, 1, bulkJobId).Err()
		if onCompleteJobId != "" {
			jp.EnqueueJobId(onCompleteJobId)
		}
	} else {
		// put the job back into the pending queue
		jp.rdb.LPush(ctx, pendingJobQueue, bulkJobId).Err()
	}

	return nil
}

func (jp *JobProcessor) processFetchPlaidTransactions(job *Job) error {
	log.Printf("🔄 Processing Plaid transactions fetch job: %s", job.ID)
	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}
	log.Printf("🔄 Job data: %v", jobData)
	accountID := jobData["account_id"].(string)
	userID := int(jobData["user_id"].(float64))
	monthYear, ok := jobData["month_year"].(int)
	if !ok {
		// Use current month if not provided
		now := time.Now()
		monthYear = int(now.Month())*10000 + now.Year()
	}
	accessToken, isSandbox, err := database.GetAccessTokenFromAccountID(accountID)
	if err != nil {
		return fmt.Errorf("failed to get access token from account id: %w", err)
	}
	// Set startDate to the first day of the month, endDate to the last day of the month
	month := int(monthYear / 10000)
	year := int(monthYear % 10000)
	location := time.Now().Location()
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, location).Format(time.RFC3339)
	endDate := time.Date(year, time.Month(month+1), 0, 23, 59, 59, 0, location).Format(time.RFC3339)
	log.Printf("🔄 Fetching transactions from %s to %s", startDate, endDate)
	transactions, err := plaid.GetTransactions(accessToken, startDate, endDate, isSandbox)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}
	log.Printf("✅ Fetched %d transactions from Plaid", len(transactions))
	// save transactions to database
	err = database.CreatePlaidTransactions(userID, accountID, transactions)
	if err != nil {
		return fmt.Errorf("failed to create plaid transactions: %w", err)
	}
	log.Printf("✅ Created %d transactions", len(transactions))
	// Mark plaid Account as synced
	err = database.MarkPlaidAccountAsSynced(accountID)
	if err != nil {
		return fmt.Errorf("failed to mark plaid account as synced: %w", err)
	}
	log.Printf("✅ Completed Plaid transactions fetch job: %s", job.ID)
	return nil
}

func (jp *JobProcessor) processDailyBalance(job *Job) error {
	log.Printf("🔄 Processing daily balance job: %s", job.ID)
	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}
	log.Printf("🔄 Job data: %v", jobData)
	userID := int(jobData["user_id"].(float64))
	monthYear := int(jobData["month_year"].(float64))

	err := database.ProcessDailyBalance(userID, monthYear)
	if err != nil {
		log.Printf("❌ Worker %s: Error dequeuing job: %v", job.ID, err)
		return err
	}
	log.Printf("✅ Finished daily balance job: %s", job.ID)
	return nil
}

// processJobMonitor monitors the progress of multiple jobs
func (jp *JobProcessor) processJobMonitor(job *Job) error {
	log.Printf("🔄 Processing job monitor: %s", job.ID)

	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	batchID := jobData["batch_id"].(string)
	jobs, err := jp.rdb.LRange(ctx, batchID, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get jobs: %w", err)
	}

	for _, jobID := range jobs {
		jobStatus, err := jp.GetJobStatusByID(jobID)
		if err != nil {
			return fmt.Errorf("failed to get job: %w", err)
		}
		if jobStatus == "completed" {
			jp.rdb.LRem(ctx, batchID, 1, jobID)
		}
	}

	return nil
}

// StartWorker starts a single background worker
func (jp *JobProcessor) StartWorker(workerID int) {
	log.Printf("🚀 Starting worker %d...", workerID)

	for {
		job, err := jp.DequeueJob()
		if err != nil {
			log.Printf("❌ Worker %d: Error dequeuing job: %v", workerID, err)
			continue
		}

		if job == nil {
			// No jobs available, continue polling
			continue
		}

		// Process the job
		log.Printf("🔄 Worker %d: Processing job: %s (Type: %s)", workerID, job.ID, job.Type)
		err = jp.ProcessJob(job)
		if err != nil {
			log.Printf("❌ Worker %d: Error processing job %s: %v", workerID, job.ID, err)
		}

		// Clean up completed jobs from processing queue
		go jp.cleanupCompletedJobs()
	}
}

// cleanupCompletedJobs removes completed jobs from the processing queue
func (jp *JobProcessor) cleanupCompletedJobs() {
	// Get all jobs from processing queue
	jobs, err := jp.rdb.LRange(ctx, "processing_queue", 0, -1).Result()
	if err != nil {
		log.Printf("Warning: failed to get processing queue: %v", err)
		return
	}

	for _, jobJSON := range jobs {
		var job Job
		if err := json.Unmarshal([]byte(jobJSON), &job); err != nil {
			log.Printf("Warning: failed to unmarshal job: %v", err)
			continue
		}

		// Check if job is completed or failed
		status, err := jp.GetJobStatus(job.ID)
		if err != nil {
			log.Printf("Warning: failed to get status for job %s: %v", job.ID, err)
			continue
		}

		// Remove completed/failed jobs from processing queue
		if status == "completed" || status == "failed" {
			removed, err := jp.rdb.LRem(ctx, "processing_queue", 1, jobJSON).Result()
			if err != nil {
				log.Printf("Warning: failed to remove job %s from processing queue: %v", job.ID, err)
				continue
			}

			// Only log if we actually removed the job
			if removed > 0 {
				log.Printf("🧹 Cleaned up %s job %s from processing queue", status, job.ID)
			}
		}
	}
}

// StartWorkers starts multiple background workers
func (jp *JobProcessor) StartWorkers(numWorkers int) {
	log.Printf("🚀 Starting %d background workers...", numWorkers)

	for i := 1; i <= numWorkers; i++ {
		go jp.StartWorker(i)
	}
}

func (jp *JobProcessor) StartBulkJobWorker() {
	log.Printf("🚀 Starting bulk job worker...")

	for {
		jobCompleted, err := jp.DequeueOrRequeueBulkJob()
		if err != nil {
			log.Printf("❌ Worker: Error dequeuing bulk job: %v", err)
			continue
		}
		if jobCompleted {
			log.Printf("✅ Bulk job completed")
		}
		time.Sleep(1 * time.Second)
	}
}

// EnqueueSampleJobs adds some sample jobs to the queue
func (jp *JobProcessor) EnqueueSampleJobs() {
	log.Println("📤 Enqueueing sample jobs...")

	// Enqueue some hello world jobs
	job1, _ := jp.CreateJob("hello_world", json.RawMessage(`"Welcome to Redis!"`))
	jp.EnqueueJob(job1)
	job2, _ := jp.CreateJob("hello_world", json.RawMessage(`"Processing jobs in background"`))
	jp.EnqueueJob(job2)
	job3, _ := jp.CreateJob("hello_world", json.RawMessage(`"Redis queue is awesome"`))
	jp.EnqueueJob(job3)

	// Enqueue some print message jobs
	job4, _ := jp.CreateJob("print_message", json.RawMessage(`"This is a test message"`))
	jp.EnqueueJob(job4)
	job5, _ := jp.CreateJob("print_message", json.RawMessage(`"Background processing works!"`))
	jp.EnqueueJob(job5)
	job6, _ := jp.CreateJob("print_message", json.RawMessage(`"Redis + Go = ❤️"`))
	jp.EnqueueJob(job6)

	log.Println("✅ Sample jobs enqueued!")
}

// HTTP handlers
func (jp *JobProcessor) handleEnqueueJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		http.Error(w, "Job type is required", http.StatusBadRequest)
		return
	}

	// Enqueue job using the new status-tracking system
	job, err := jp.CreateJob(req.Type, req.Data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create job: %v", err), http.StatusInternalServerError)
		return
	}
	err = jp.EnqueueJob(job)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to enqueue job: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	response := EnqueueResponse{
		Success: true,
		JobID:   "job_enqueued", // We'll get the actual ID from the EnqueueJob function
		Message: "Job enqueued successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetJobStatus returns the status of a specific job
func (jp *JobProcessor) handleGetJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	job, err := jp.GetJob(jobID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get job: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// handleCreateJobMonitor creates a monitoring job for multiple jobs
func (jp *JobProcessor) handleCreateJobMonitor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JobIDs      []string `json:"job_ids"`
		Description string   `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.JobIDs) == 0 {
		http.Error(w, "At least one job ID is required", http.StatusBadRequest)
		return
	}

	// Create monitor job data
	monitorData := map[string]interface{}{
		"job_ids":     req.JobIDs,
		"description": req.Description,
	}

	monitorDataJSON, err := json.Marshal(monitorData)
	if err != nil {
		http.Error(w, "Failed to marshal monitor data", http.StatusInternalServerError)
		return
	}

	// Enqueue the monitor job
	job, err := jp.CreateJob("job_monitor", monitorDataJSON)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create monitor job: %v", err), http.StatusInternalServerError)
		return
	}
	err = jp.EnqueueJob(job)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create monitor job: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"message":     "Job monitor created successfully",
		"job_ids":     req.JobIDs,
		"description": req.Description,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (jp *JobProcessor) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// StartHTTPServer starts the HTTP server
func (jp *JobProcessor) StartHTTPServer(port string) {
	http.HandleFunc("/enqueue", jp.handleEnqueueJob)
	http.HandleFunc("/health", jp.handleHealth)
	http.HandleFunc("/job/status", jp.handleGetJobStatus)
	http.HandleFunc("/job/monitor", jp.handleCreateJobMonitor)

	log.Printf("🌐 Starting HTTP server on port %s", port)
	log.Printf("📋 Available endpoints:")
	log.Printf("   POST /enqueue        - Enqueue a new job")
	log.Printf("   GET  /health         - Health check")
	log.Printf("   GET  /job/status     - Get job status")
	log.Printf("   POST /job/monitor    - Create job monitor")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start HTTP server:", err)
	}
}

func main() {
	plaid.InitPlaid()
	// Get Redis address from environment variable
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // fallback
	}

	// Get worker port from environment variable
	workerPort := os.Getenv("WORKER_PORT")
	if workerPort == "" {
		workerPort = "8081" // fallback
	}

	// Get database connection string from environment variable
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		log.Fatal("DATABASE_URL is not set")
		// dbConnStr = "postgres://postgres:password@localhost:5432/watson?sslmode=disable" // fallback
	}

	// Initialize shared database connection
	if err := database.InitDB(dbConnStr); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.CloseDB()

	log.Println("✅ Connected to database successfully!")

	// Create job processor
	processor := NewJobProcessor(redisAddr)

	// Test Redis connection
	_, err := processor.rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("✅ Connected to Redis successfully!")

	// Enqueue some sample jobs
	processor.EnqueueSampleJobs()

	// Start 3 background workers
	processor.StartWorkers(10)
	go processor.StartBulkJobWorker()
	// Start the HTTP server
	processor.StartHTTPServer(workerPort)
}
