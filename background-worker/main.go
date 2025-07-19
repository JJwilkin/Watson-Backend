package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"watson/database"
	"watson/plaid"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// TellerAccount represents an account from the Teller API
type TellerAccount struct {
	ID                  string `json:"id"`
	TellerInstitutionID string `json:"teller_institution_id"`
	EnrollmentID        string `json:"enrollment_id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Subtype             string `json:"subtype"`
	Currency            string `json:"currency"`
	LastFour            string `json:"last_four"`
	Status              string `json:"status"`
	Institution         struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"institution"`
	Links struct {
		Self         string `json:"self"`
		Details      string `json:"details"`
		Balances     string `json:"balances"`
		Transactions string `json:"transactions"`
	} `json:"links"`
}

// TellerTransaction represents a transaction from the Teller API
type TellerTransaction struct {
	ID             string `json:"id"`
	AccountID      string `json:"account_id"`
	Amount         string `json:"amount"` // Amount is a string in Teller API
	Description    string `json:"description"`
	Date           string `json:"date"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	RunningBalance string `json:"running_balance"` // Can be null, so using string
	Details        struct {
		ProcessingStatus string `json:"processing_status"`
		Category         string `json:"category"`
		Counterparty     struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"counterparty"`
	} `json:"details"`
	Links struct {
		Self    string `json:"self"`
		Account string `json:"account"`
	} `json:"links"`
}

// Job represents a task to be processed
type Job struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"created_at"`
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

// EnqueueJob adds a job to the queue
func (jp *JobProcessor) EnqueueJob(jobType string, data json.RawMessage) error {
	job := Job{
		ID:        fmt.Sprintf("job_%d", time.Now().UnixNano()),
		Type:      jobType,
		Data:      data,
		CreatedAt: time.Now(),
	}

	jobJSON, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add job to the queue (using LPUSH to add to the left of the list)
	err = jp.rdb.LPush(ctx, "job_queue", jobJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	log.Printf("✅ Enqueued job: %s (Type: %s)", job.ID, job.Type)
	return nil
}

// DequeueJob removes and returns a job from the queue
func (jp *JobProcessor) DequeueJob() (*Job, error) {
	// Use BRPOP to block until a job is available (timeout: 5 seconds)
	result, err := jp.rdb.BRPop(ctx, 5*time.Second, "job_queue").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected result format")
	}

	jobJSON := result[1] // result[0] is the key name, result[1] is the value
	var job Job
	err = json.Unmarshal([]byte(jobJSON), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// ProcessJob handles the actual job processing
func (jp *JobProcessor) ProcessJob(job *Job) error {
	log.Printf("🔄 Processing job: %s (Type: %s)", job.ID, job.Type)

	switch job.Type {
	case "hello_world":
		return jp.processHelloWorld(job)
	case "print_message":
		return jp.processPrintMessage(job)
	case "new_teller_link":
		return jp.processTellerSuccess(job)
	case "fetch_transactions":
		return jp.processFetchTransactions(job)
	case "initial_plaid_sync":
		return jp.processInitialPlaidSync(job)
	case "fetch_plaid_transactions":
		return jp.processFetchPlaidTransactions(job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
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
	time.Sleep(500 * time.Millisecond)

	log.Printf("✅ Completed print message job: %s", job.ID)
	return nil
}

func (jp *JobProcessor) processFetchTransactions(job *Job) error {
	log.Printf("🔄 Processing fetch transactions job: %s", job.ID)

	// Parse job data to get account ID and user ID
	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	transactions_link, ok := jobData["transactions_link"].(string)
	if !ok {
		return fmt.Errorf("transactions_link not found in job data")
	}

	access_token, ok := jobData["access_token"].(string)
	if !ok {
		return fmt.Errorf("access_token not found in job data")
	}

	user_id, ok := jobData["user_id"].(float64) // JSON numbers are unmarshaled as float64
	if !ok {
		return fmt.Errorf("user_id not found in job data")
	}

	teller_institution_id, ok := jobData["teller_institution_id"].(string)
	if !ok {
		return fmt.Errorf("teller_institution_id not found in job data")
	}

	account_id, ok := jobData["account_id"].(string)
	if !ok {
		return fmt.Errorf("account_id not found in job data")
	}

	transactions, err := jp.fetchTellerTransactions(transactions_link, access_token)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	// Save all transactions to the database in a single batch
	savedTransactions, err := jp.SaveTellerTransactions(int(user_id), teller_institution_id, account_id, transactions)
	if err != nil {
		return fmt.Errorf("failed to save transactions: %w", err)
	}

	log.Printf("✅ Fetched and saved %d transactions for account: %s", len(savedTransactions), transactions_link)
	return nil
}

func (jp *JobProcessor) fetchTellerTransactions(transactions_link string, access_token string) ([]TellerTransaction, error) {
	log.Printf("🔄 Fetching Teller transactions for link: %s", transactions_link)

	// Create request to Teller API
	req, err := http.NewRequest("GET", transactions_link, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic auth (username is access token, password is empty)
	req.SetBasicAuth(access_token, "")

	// Make the request
	resp, err := jp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse accounts from response
	var transactions []TellerTransaction
	if err := json.Unmarshal(body, &transactions); err != nil {
		return nil, fmt.Errorf("failed to parse transactions response: %w", err)
	}

	log.Printf("✅ Successfully fetched %d transactions from Teller API", len(transactions))
	return transactions, nil
}

// processTellerSuccess handles Teller success jobs
func (jp *JobProcessor) processTellerSuccess(job *Job) error {
	log.Printf("🔄 Processing Teller success job: %s", job.ID)

	// Parse job data to get access token and user ID
	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(job.Data), &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	// Extract access token and user ID from job data
	accessToken, ok := jobData["token"].(string)
	if !ok {
		return fmt.Errorf("access token not found in job data")
	}

	userID, ok := jobData["user_id"].(float64) // JSON numbers are unmarshaled as float64
	if !ok {
		return fmt.Errorf("user_id not found in job data")
	}

	// Call the Teller API to fetch accounts
	accounts, err := jp.fetchTellerAccounts(accessToken)
	if err != nil {
		return fmt.Errorf("failed to fetch Teller accounts: %w", err)
	}

	createdAccounts := []TellerAccount{}
	// Save each account to the database
	for _, account := range accounts {
		savedAccount, err := jp.SaveTellerAccount(int(userID), accessToken, account)
		if err != nil {
			log.Printf("❌ Failed to save account %s: %v", account.ID, err)
			continue
		}
		log.Printf("✅ Saved account: %s (%s) - %s", savedAccount.Name, savedAccount.Type, savedAccount.Institution.Name)
		createdAccounts = append(createdAccounts, *savedAccount)
	}

	log.Printf("✅ Completed Teller success job: %s", job.ID)
	// enqueue job to fetch transactions for each teller account
	for _, account := range createdAccounts {
		// Create JSON data for the fetch_transactions job
		jobData := map[string]interface{}{
			"account_id":            account.ID,
			"user_id":               int(userID),
			"access_token":          accessToken,
			"transactions_link":     account.Links.Transactions,
			"teller_institution_id": account.TellerInstitutionID,
		}
		jobDataJSON, _ := json.Marshal(jobData)
		jp.EnqueueJob("fetch_transactions", jobDataJSON)
	}
	return nil
}

// SaveTellerTransactions saves multiple Teller transactions to the database in a single batch
func (jp *JobProcessor) SaveTellerTransactions(userID int, teller_institution_id string, teller_account_id string, transactions []TellerTransaction) ([]TellerTransaction, error) {
	if len(transactions) == 0 {
		return []TellerTransaction{}, nil
	}

	// Start a transaction for batch insert
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() is called

	// Prepare the batch insert statement
	query := `
		INSERT INTO transactions (
			user_id, teller_institution_id, teller_account_id, teller_transaction_id,
			amount, description, date, type, status, running_balance,
			processing_status, category, counterparty_name, counterparty_type,
			self_link, account_link, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) ON CONFLICT (id) DO UPDATE SET
			amount = EXCLUDED.amount,
			description = EXCLUDED.description,
			date = EXCLUDED.date,
			type = EXCLUDED.type,
			status = EXCLUDED.status,
			running_balance = EXCLUDED.running_balance,
			processing_status = EXCLUDED.processing_status,
			category = EXCLUDED.category,
			counterparty_name = EXCLUDED.counterparty_name,
			counterparty_type = EXCLUDED.counterparty_type,
			self_link = EXCLUDED.self_link,
			account_link = EXCLUDED.account_link,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, teller_transaction_id, amount, description, date, type, status, running_balance,
			processing_status, category, counterparty_name, counterparty_type, self_link, account_link, created_at, updated_at
	`

	// Prepare the statement
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert all transactions
	var savedTransactions []TellerTransaction
	for _, transaction := range transactions {
		var savedTransaction TellerTransaction
		var dbID string
		var dbUserID int
		var createdAt, updatedAt time.Time

		err := stmt.QueryRow(
			userID, teller_institution_id, teller_account_id, transaction.ID,
			transaction.Amount, transaction.Description, transaction.Date, transaction.Type, transaction.Status, transaction.RunningBalance,
			transaction.Details.ProcessingStatus, transaction.Details.Category, transaction.Details.Counterparty.Name, transaction.Details.Counterparty.Type,
			transaction.Links.Self, transaction.Links.Account,
		).Scan(
			&dbID, &dbUserID, &savedTransaction.ID, &savedTransaction.Amount, &savedTransaction.Description,
			&savedTransaction.Date, &savedTransaction.Type, &savedTransaction.Status, &savedTransaction.RunningBalance,
			&savedTransaction.Details.ProcessingStatus, &savedTransaction.Details.Category,
			&savedTransaction.Details.Counterparty.Name, &savedTransaction.Details.Counterparty.Type,
			&savedTransaction.Links.Self, &savedTransaction.Links.Account, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert transaction %s: %w", transaction.ID, err)
		}

		savedTransaction.AccountID = transaction.AccountID
		savedTransactions = append(savedTransactions, savedTransaction)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✅ Successfully saved %d transactions to database", len(savedTransactions))
	return savedTransactions, nil
}

// SaveTellerAccount saves a Teller account to the database and returns the saved account
func (jp *JobProcessor) SaveTellerAccount(userID int, accessToken string, account TellerAccount) (*TellerAccount, error) {
	query := `
		INSERT INTO teller_accounts (
			id, user_id, teller_institution_id, enrollment_id, 
			account_name, account_type, account_subtype, currency, last_four, status,
			institution_id, institution_name, self_link, details_link, balances_link, transactions_link
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) ON CONFLICT (id) DO UPDATE SET
			account_name = EXCLUDED.account_name,
			account_type = EXCLUDED.account_type,
			account_subtype = EXCLUDED.account_subtype,
			currency = EXCLUDED.currency,
			last_four = EXCLUDED.last_four,
			status = EXCLUDED.status,
			institution_name = EXCLUDED.institution_name,
			self_link = EXCLUDED.self_link,
			details_link = EXCLUDED.details_link,
			balances_link = EXCLUDED.balances_link,
			transactions_link = EXCLUDED.transactions_link,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, account_name, account_type, account_subtype, institution_name, self_link, details_link, balances_link, transactions_link,
			created_at, updated_at
	`

	// Get the teller_institution_id for this user and enrollment
	var tellerInstitutionID string
	err := database.DB.QueryRow(
		"SELECT id FROM teller_institutions WHERE user_id = $1 AND access_token = $2",
		userID, accessToken,
	).Scan(&tellerInstitutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teller institution ID: %w", err)
	}

	// Execute the query and scan the returned data
	var savedAccount TellerAccount
	var dbUserID int
	var createdAt, updatedAt time.Time

	err = database.DB.QueryRow(query,
		account.ID, userID, tellerInstitutionID, account.EnrollmentID,
		account.Name, account.Type, account.Subtype, account.Currency, account.LastFour, account.Status,
		account.Institution.ID, account.Institution.Name,
		account.Links.Self, account.Links.Details, account.Links.Balances, account.Links.Transactions,
	).Scan(
		&savedAccount.ID, &dbUserID, &savedAccount.Name, &savedAccount.Type, &savedAccount.Subtype, &savedAccount.Institution.Name, &savedAccount.Links.Self, &savedAccount.Links.Details, &savedAccount.Links.Balances, &savedAccount.Links.Transactions,
		&createdAt, &updatedAt,
	)

	savedAccount.TellerInstitutionID = tellerInstitutionID

	if err != nil {
		return nil, fmt.Errorf("failed to save teller account: %w", err)
	}

	return &savedAccount, nil
}

// fetchTellerAccounts fetches accounts from Teller API using client certificates
func (jp *JobProcessor) fetchTellerAccounts(accessToken string) ([]TellerAccount, error) {
	log.Printf("🔄 Fetching Teller accounts for token: %.10s...", accessToken)

	// Create request to Teller API
	req, err := http.NewRequest("GET", "https://api.teller.io/accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic auth (username is access token, password is empty)
	req.SetBasicAuth(accessToken, "")

	// Make the request
	resp, err := jp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse accounts from response
	var accounts []TellerAccount
	if err := json.Unmarshal(body, &accounts); err != nil {
		return nil, fmt.Errorf("failed to parse accounts response: %w", err)
	}

	log.Printf("✅ Successfully fetched %d accounts from Teller API", len(accounts))
	return accounts, nil
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
	accounts, err := plaid.GetAccounts(accessToken)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	log.Printf("✅ Fetched %d accounts from Plaid", len(accounts))
	plaidTokenID, userID, err := database.GetUserIdFromAccessToken(accessToken)
	if err != nil {
		return fmt.Errorf("failed to get user id from access token: %w", err)
	}
	err = database.CreatePlaidAccount(userID, plaidTokenID, accounts)
	if err != nil {
		return fmt.Errorf("failed to create plaid account: %w", err)
	}
	err = database.MarkPlaidTokenAsProcessed(plaidTokenID)
	if err != nil {
		return fmt.Errorf("failed to mark plaid token as processed: %w", err)
	}
	log.Printf("✅ Completed initial Plaid sync job: %s", job.ID)

	// enqueue job to fetch transactions for each plaid account
	for _, account := range accounts {
		jobData := map[string]interface{}{
			"account_id": account.GetAccountId(),
			"user_id":    userID,
		}
		jobDataJSON, _ := json.Marshal(jobData)
		jp.EnqueueJob("fetch_plaid_transactions", jobDataJSON)
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
	accessToken, err := database.GetAccessTokenFromAccountID(accountID)
	if err != nil {
		return fmt.Errorf("failed to get access token from account id: %w", err)
	}
	// FAKE START AND END DATE FOR NOW
	startDate := time.Now().Add(-365 * 24 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().Format(time.RFC3339)
	log.Printf("🔄 Fetching transactions from %s to %s", startDate, endDate)
	transactions, err := plaid.GetTransactions(accessToken, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}
	log.Printf("✅ Fetched %d transactions from Plaid", len(transactions))
	// save transactions to database
	err = database.CreatePlaidTransactions(userID, accountID, transactions)
	if err != nil {
		return fmt.Errorf("failed to create plaid transactions: %w", err)
	}
	log.Printf("✅ Completed Plaid transactions fetch job: %s", job.ID)
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
	}
}

// StartWorkers starts multiple background workers
func (jp *JobProcessor) StartWorkers(numWorkers int) {
	log.Printf("🚀 Starting %d background workers...", numWorkers)

	for i := 1; i <= numWorkers; i++ {
		go jp.StartWorker(i)
	}
}

// EnqueueSampleJobs adds some sample jobs to the queue
func (jp *JobProcessor) EnqueueSampleJobs() {
	log.Println("📤 Enqueueing sample jobs...")

	// Enqueue some hello world jobs
	jp.EnqueueJob("hello_world", json.RawMessage(`"Welcome to Redis!"`))
	jp.EnqueueJob("hello_world", json.RawMessage(`"Processing jobs in background"`))
	jp.EnqueueJob("hello_world", json.RawMessage(`"Redis queue is awesome"`))

	// Enqueue some print message jobs
	jp.EnqueueJob("print_message", json.RawMessage(`"This is a test message"`))
	jp.EnqueueJob("print_message", json.RawMessage(`"Background processing works!"`))
	jp.EnqueueJob("print_message", json.RawMessage(`"Redis + Go = ❤️"`))

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

	// Create job
	job := Job{
		ID:        fmt.Sprintf("job_%d", time.Now().UnixNano()),
		Type:      req.Type,
		Data:      req.Data,
		CreatedAt: time.Now(),
	}

	jobJSON, err := json.Marshal(job)
	if err != nil {
		http.Error(w, "Failed to marshal job", http.StatusInternalServerError)
		return
	}

	// Enqueue job
	err = jp.rdb.LPush(ctx, "job_queue", jobJSON).Err()
	if err != nil {
		http.Error(w, "Failed to enqueue job", http.StatusInternalServerError)
		return
	}

	// Return response
	response := EnqueueResponse{
		Success: true,
		JobID:   job.ID,
		Message: fmt.Sprintf("Job enqueued successfully: %s", job.ID),
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

	log.Printf("🌐 Starting HTTP server on port %s", port)
	log.Printf("📋 Available endpoints:")
	log.Printf("   POST /enqueue - Enqueue a new job")
	log.Printf("   GET  /health  - Health check")

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
	processor.StartWorkers(3)

	// Start the HTTP server
	processor.StartHTTPServer(workerPort)
}
