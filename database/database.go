package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	plaid "github.com/plaid/plaid-go/v31/plaid"
)

// Database connection
var DB *sql.DB

// User represents a user in the database
type DBUser struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"-"` // Don't expose password in JSON
}

// Transaction represents a transaction in the database
type Transaction struct {
	TransactionID   string    `json:"transaction_id"`
	UserID          int       `json:"user_id"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	TransactionDate time.Time `json:"transaction_date"`
	Category        string    `json:"category"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	Type            string    `json:"type"`
	ProviderType    string    `json:"provider_type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TellerInstitution struct {
	ID          string `json:"id"`
	UserID      int    `json:"user_id"`
	Name        string `json:"name"`
	TellerID    string `json:"teller_id"`
	AccessToken string `json:"access_token"`
}

// TellerPayload represents the structure of the Teller webhook payload
type TellerPayload struct {
	AccessToken string `json:"accessToken"`
	User        struct {
		ID string `json:"id"`
	} `json:"user"`
	Enrollment struct {
		ID          string `json:"id"`
		Institution struct {
			Name string `json:"name"`
		} `json:"institution"`
	} `json:"enrollment"`
	Signatures []string `json:"signatures"`
}

type MonthlySummary struct {
	ID                     int       `json:"id"`
	UserID                 int       `json:"user_id"`
	MonthYear              int       `json:"monthyear"`
	TotalSpent             float64   `json:"total_spent"`
	FixedExpenses          float64   `json:"fixed_expenses"`
	SavingTargetPercentage float64   `json:"saving_target_percentage"`
	StartingBalance        float64   `json:"starting_balance"`
	Income                 float64   `json:"income"`
	SavedAmount            float64   `json:"saved_amount"`
	Invested               float64   `json:"invested"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type MonthlyBudgetSpendCategory struct {
	ID               string    `json:"id"`
	UserID           int       `json:"user_id"`
	MonthlySummaryID int       `json:"monthly_summary_id"`
	MonthYear        int       `json:"monthyear"`
	Category         string    `json:"category"`
	Budget           float64   `json:"budget"`
	DailyAllowance   float64   `json:"daily_allowance"`
	TotalSpent       float64   `json:"total_spent"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Monthly Balance
type MonthlyBalance struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	MonthYear        int       `json:"monthyear"`
	TotalOwing       float64   `json:"total_owing"`
	NetCash          float64   `json:"net_cash"`
	AvailableBalance float64   `json:"available_balance"`
	CurrentBalance   float64   `json:"current_balance"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Saving Goals
type SavingsGoal struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Name         string    `json:"name"`
	TotalAmount  float64   `json:"total_amount"`
	Redeemed     bool      `json:"redeemed"`
	CurrentSaved float64   `json:"current_saved"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// InitDB initializes the database connection
func InitDB(connStr string) error {
	log.Printf("Connecting to database: %s", connStr)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connection established successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

// CreateUser creates a new user in the database
func CreateUser(email, password string) (*DBUser, error) {
	var user DBUser
	query := "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING user_id, email"

	err := DB.QueryRow(query, email, password).Scan(&user.UserID, &user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return &user, nil
}

func GetUserByEmailAndPassword(email, password string) (*DBUser, error) {
	var user DBUser
	query := "SELECT user_id, email, password FROM users WHERE email = $1 AND password = $2"

	err := DB.QueryRow(query, email, password).Scan(&user.UserID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*DBUser, error) {
	var user DBUser
	query := "SELECT user_id, email, password FROM users WHERE email = $1"

	err := DB.QueryRow(query, email).Scan(&user.UserID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(userID int) (*DBUser, error) {
	var user DBUser
	query := "SELECT user_id, email, password FROM users WHERE user_id = $1"

	err := DB.QueryRow(query, userID).Scan(&user.UserID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

func CreateTellerInstitution(userID int, name string, tellerID string, accessToken string) (*TellerInstitution, error) {
	var tellerInstitution TellerInstitution
	query := "INSERT INTO teller_institutions (user_id, name, teller_id, access_token) VALUES ($1, $2, $3, $4) RETURNING id, user_id, name, teller_id, access_token"
	err := DB.QueryRow(query, userID, name, tellerID, accessToken).Scan(&tellerInstitution.ID, &tellerInstitution.UserID, &tellerInstitution.Name, &tellerInstitution.TellerID, &tellerInstitution.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create teller institution: %v", err)
	}
	return &tellerInstitution, nil
}

// HandleTellerSuccess processes the Teller webhook payload and creates a TellerInstitution record
func HandleTellerSuccess(userID int, payload TellerPayload) (*TellerInstitution, error) {
	// Create the teller institution record
	tellerInstitution, err := CreateTellerInstitution(
		userID,
		payload.Enrollment.Institution.Name,
		payload.Enrollment.ID, // Using enrollment ID as teller_id
		payload.AccessToken,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create teller institution: %v", err)
	}

	return tellerInstitution, nil
}

// CreateTransaction creates a new transaction
func CreateTransaction(userID int, description string, amount float64, transactionDate time.Time) (*Transaction, error) {
	var transaction Transaction
	query := `
		INSERT INTO transactions (user_id, description, amount, transaction_date) 
		VALUES ($1, $2, $3, $4) 
		RETURNING transaction_id, user_id, description, amount, transaction_date, created_at, updated_at
	`

	err := DB.QueryRow(query, userID, description, amount, transactionDate).Scan(
		&transaction.TransactionID, &transaction.UserID, &transaction.Description,
		&transaction.Amount, &transaction.TransactionDate, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	return &transaction, nil
}

// GetTransactionsByUserID retrieves all transactions for a user
func GetTransactionsByUserID(userID int) ([]Transaction, error) {
	query := `
		SELECT transaction_id, user_id, description, amount, transaction_date, created_at, updated_at
		FROM transactions 
		WHERE user_id = $1 
		ORDER BY transaction_date DESC, created_at DESC
	`

	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %v", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(
			&transaction.TransactionID, &transaction.UserID, &transaction.Description,
			&transaction.Amount, &transaction.TransactionDate, &transaction.CreatedAt, &transaction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %v", err)
		}
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %v", err)
	}

	return transactions, nil
}

// GetTransactionByID retrieves a specific transaction
func GetTransactionByID(transactionID int) (*Transaction, error) {
	var transaction Transaction
	query := `
		SELECT transaction_id, user_id, description, amount, transaction_date, created_at, updated_at
		FROM transactions 
		WHERE transaction_id = $1
	`

	err := DB.QueryRow(query, transactionID).Scan(
		&transaction.TransactionID, &transaction.UserID, &transaction.Description,
		&transaction.Amount, &transaction.TransactionDate, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to get transaction: %v", err)
	}

	return &transaction, nil
}

// UpdateTransaction updates an existing transaction
func UpdateTransaction(transactionID int, description string, amount float64, transactionDate time.Time) (*Transaction, error) {
	var transaction Transaction
	query := `
		UPDATE transactions 
		SET description = $2, amount = $3, transaction_date = $4
		WHERE transaction_id = $1
		RETURNING transaction_id, user_id, description, amount, transaction_date, created_at, updated_at
	`

	err := DB.QueryRow(query, transactionID, description, amount, transactionDate).Scan(
		&transaction.TransactionID, &transaction.UserID, &transaction.Description,
		&transaction.Amount, &transaction.TransactionDate, &transaction.CreatedAt, &transaction.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to update transaction: %v", err)
	}

	return &transaction, nil
}

// DeleteTransaction deletes a transaction
func DeleteTransaction(transactionID int) error {
	query := "DELETE FROM transactions WHERE transaction_id = $1"

	result, err := DB.Exec(query, transactionID)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// GetTransactionStats returns basic statistics for a user's transactions
func GetTransactionStats(userID int) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_transactions,
			SUM(amount) as total_amount,
			AVG(amount) as average_amount,
			MIN(amount) as min_amount,
			MAX(amount) as max_amount
		FROM transactions 
		WHERE user_id = $1
	`

	var stats struct {
		TotalTransactions int     `db:"total_transactions"`
		TotalAmount       float64 `db:"total_amount"`
		AverageAmount     float64 `db:"average_amount"`
		MinAmount         float64 `db:"min_amount"`
		MaxAmount         float64 `db:"max_amount"`
	}

	err := DB.QueryRow(query, userID).Scan(
		&stats.TotalTransactions, &stats.TotalAmount, &stats.AverageAmount,
		&stats.MinAmount, &stats.MaxAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction stats: %v", err)
	}

	return map[string]interface{}{
		"total_transactions": stats.TotalTransactions,
		"total_amount":       stats.TotalAmount,
		"average_amount":     stats.AverageAmount,
		"min_amount":         stats.MinAmount,
		"max_amount":         stats.MaxAmount,
	}, nil
}

// ********** PLAID **********

func CreatePlaidToken(userID int, accessToken string, itemID string) error {
	query := "INSERT INTO plaid_tokens (user_id, access_token, item_id) VALUES ($1, $2, $3)"
	_, err := DB.Exec(query, userID, accessToken, itemID)
	if err != nil {
		return fmt.Errorf("failed to create plaid token: %v", err)
	}
	return nil
}

func MarkPlaidAccountAsSynced(accountID string) error {
	query := "UPDATE plaid_accounts SET is_processed = TRUE WHERE id = $1"
	_, err := DB.Exec(query, accountID)
	if err != nil {
		return fmt.Errorf("failed to mark plaid account as synced: %v", err)
	}
	log.Printf("Marked plaid account as synced: %s", accountID)
	return nil
}

func GetAccessTokenFromAccountID(accountId string) (string, error) {
	query := "SELECT p.access_token FROM plaid_accounts as a JOIN plaid_tokens as p ON a.plaid_token_id = p.id WHERE a.id = $1"
	var accessToken string
	err := DB.QueryRow(query, accountId).Scan(&accessToken)
	if err != nil {
		return "", fmt.Errorf("failed to get access token from account id: %v", err)
	}
	return accessToken, nil
}

func GetUserIdFromAccessToken(accessToken string) (string, int, error) {
	query := "SELECT id, user_id, is_processed FROM plaid_tokens WHERE access_token = $1"
	var plaidTokenID string
	var userID int
	var isProcessed bool
	err := DB.QueryRow(query, accessToken).Scan(&plaidTokenID, &userID, &isProcessed)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get user id from access token: %v", err)
	}
	if isProcessed {
		return "", 0, fmt.Errorf("access token already processed")
	}
	return plaidTokenID, userID, nil
}

func MarkPlaidTokenAsProcessed(plaidTokenID string) error {
	query := "UPDATE plaid_tokens SET is_processed = TRUE WHERE id = $1"
	_, err := DB.Exec(query, plaidTokenID)
	if err != nil {
		return fmt.Errorf("failed to mark plaid token as processed: %v", err)
	}
	return nil
}

func GetAllAccountsSynced(userID int) (bool, error) {
	query := "SELECT COUNT(*) FROM plaid_accounts WHERE user_id = $1 AND is_processed = FALSE"
	var count int
	err := DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to get all accounts synced: %v", err)
	}
	return count == 0, nil
}

func GetPlaidAccountsByUserID(userID int) ([]string, error) {
	query := "SELECT id FROM plaid_accounts WHERE user_id = $1"
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plaid accounts by user id: %v", err)
	}
	defer rows.Close()
	var accounts []string
	for rows.Next() {
		var accountID string
		err := rows.Scan(&accountID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plaid account: %v", err)
		}
		accounts = append(accounts, accountID)
	}
	return accounts, nil
}

func CreatePlaidAccount(userID int, plaidTokenID string, accounts []plaid.AccountBase) error {
	if len(accounts) == 0 {
		return nil
	}

	// Build bulk insert query
	query := "INSERT INTO plaid_accounts (id, user_id, plaid_token_id, available_balance, current_balance, currency, account_name, official_name, account_type, account_subtype) VALUES "

	values := make([]interface{}, 0, len(accounts)*10)
	placeholders := make([]string, 0, len(accounts))

	for i, account := range accounts {
		// Create placeholder string for this account
		start := i * 10
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			start+1, start+2, start+3, start+4, start+5, start+6, start+7, start+8, start+9, start+10))

		// Extract values from Plaid account
		var availableBalance, currentBalance float64
		if account.Balances.Available.Get() != nil {
			availableBalance = *account.Balances.Available.Get()
		}
		if account.Balances.Current.Get() != nil {
			currentBalance = *account.Balances.Current.Get()
		}

		values = append(values,
			account.GetAccountId(),
			userID,
			plaidTokenID,
			availableBalance,
			currentBalance,
			account.Balances.GetIsoCurrencyCode(),
			account.GetName(),
			account.GetOfficialName(),
			string(account.GetType()),
			string(account.GetSubtype()),
		)
	}

	query += strings.Join(placeholders, ", ")
	query += " ON CONFLICT (id) DO UPDATE SET " +
		"user_id = EXCLUDED.user_id, " +
		"plaid_token_id = EXCLUDED.plaid_token_id, " +
		"available_balance = EXCLUDED.available_balance, " +
		"current_balance = EXCLUDED.current_balance, " +
		"currency = EXCLUDED.currency, " +
		"account_name = EXCLUDED.account_name, " +
		"official_name = EXCLUDED.official_name, " +
		"account_type = EXCLUDED.account_type, " +
		"account_subtype = EXCLUDED.account_subtype"

	_, err := DB.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to upsert plaid accounts: %v", err)
	}

	return nil
}

func CreatePlaidTransactions(userID int, accountID string, transactions []plaid.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	// Build bulk insert query
	query := "INSERT INTO transactions (user_id, plaid_account_id, plaid_transaction_id, amount, date, description, category, currency, status, type, provider_type) VALUES "

	values := make([]interface{}, 0, len(transactions)*11)
	placeholders := make([]string, 0, len(transactions))

	for i, transaction := range transactions {
		start := i * 11
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			start+1, start+2, start+3, start+4, start+5, start+6, start+7, start+8, start+9, start+10, start+11))

		// Handle category array - convert to JSONB format
		categorySlice := transaction.GetCategory()
		var category interface{}
		if len(categorySlice) == 0 {
			category = "[]" // Empty array as JSON string
		} else {
			// Convert slice to JSON string for JSONB insertion
			categoryJSON, err := json.Marshal(categorySlice)
			if err != nil {
				return fmt.Errorf("failed to marshal category: %v", err)
			}
			category = string(categoryJSON)
		}

		var status string
		if transaction.GetPending() {
			status = "pending"
		} else {
			status = "posted"
		}

		values = append(values,
			userID,
			accountID,
			transaction.GetTransactionId(),
			transaction.GetAmount(),
			transaction.GetDate(),
			transaction.GetName(),
			category,
			transaction.GetIsoCurrencyCode(),
			status,
			transaction.GetPaymentChannel(),
			"plaid",
		)
	}

	query += strings.Join(placeholders, ", ")
	_, err := DB.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to upsert plaid transactions: %v", err)
	}

	return nil
}

// ********** MONTHLY SUMMARY **********

func GetTransactionsExcludingCategories(userID int, categoriesToExclude []string, monthYear int) ([]Transaction, error) {
	year := monthYear % 10000
	month := monthYear / 10000
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	log.Printf("Getting transactions for user %d, categories to exclude %v, month %d", userID, categoriesToExclude, monthYear)
	log.Printf("Start date: %s, End date: %s", startDate, endDate)
	// Build the query to exclude transactions that contain any of the specified categories
	query := "SELECT id, user_id, amount, date, description, category, currency, status, type, provider_type FROM transactions WHERE user_id = $1 AND date BETWEEN $2 AND $3"

	var rows *sql.Rows
	var err error

	// If there are categories to exclude, add the exclusion condition
	if len(categoriesToExclude) > 0 {
		query += " AND NOT (category ?| $4::text[])"
		// Convert []string to pq.StringArray for PostgreSQL
		rows, err = DB.Query(query, userID, startDate, endDate, pq.Array(categoriesToExclude))
	} else {
		// If no categories to exclude, just get all transactions
		rows, err = DB.Query(query, userID, startDate, endDate)
	}
	if err != nil {
		log.Printf("Failed to query transactions: %v", err)
		return nil, fmt.Errorf("failed to query transactions: %v", err)
	}
	defer rows.Close()
	var transactions []Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.TransactionID, &transaction.UserID, &transaction.Amount, &transaction.TransactionDate, &transaction.Description, &transaction.Category, &transaction.Currency, &transaction.Status, &transaction.Type, &transaction.ProviderType)
		if err != nil {
			log.Printf("Failed to scan transaction: %v", err)
			return nil, fmt.Errorf("failed to scan transaction: %v", err)
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func GetTransactionsByCategory(userID int, category string, monthYear int) ([]Transaction, error) {

	year := monthYear % 10000
	month := monthYear / 10000
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	log.Printf("Getting transactions for user %d, category %s, month %d", userID, category, monthYear)
	log.Printf("Start date: %s, End date: %s", startDate, endDate)
	var rows *sql.Rows
	var err error
	// if category == "general" {
	// 	query := "SELECT id, user_id, amount, date, description, category, currency, status, type, provider_type FROM transactions WHERE user_id = $1 AND date BETWEEN $2 AND $3"
	// 	rows, err = DB.Query(query, userID, startDate, endDate)
	// 	if err != nil {
	// 		log.Printf("Failed to query transactions: %v", err)
	// 		return nil, fmt.Errorf("failed to query transactions: %v", err)
	// 	}
	// 	defer rows.Close()
	// } else {
	// Create the JSON array string properly
	categoryJSON := fmt.Sprintf(`["%s"]`, category)
	query := "SELECT id, user_id, amount, date, description, category, currency, status, type, provider_type FROM transactions WHERE user_id = $1 AND date BETWEEN $2 AND $3 AND category @> $4::jsonb"
	rows, err = DB.Query(query, userID, startDate, endDate, categoryJSON)
	if err != nil {
		log.Printf("Failed to query transactions: %v", err)
		return nil, fmt.Errorf("failed to query transactions: %v", err)
	}
	defer rows.Close()
	// }
	var transactions []Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.TransactionID, &transaction.UserID, &transaction.Amount, &transaction.TransactionDate, &transaction.Description, &transaction.Category, &transaction.Currency, &transaction.Status, &transaction.Type, &transaction.ProviderType)
		if err != nil {
			log.Printf("Failed to scan transaction: %v", err)
			return nil, fmt.Errorf("failed to scan transaction: %v", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

//	func GetOrCreateMonthlySummary(userID int, monthYear int) (*MonthlySummary, error) {
//		monthlySummary, _ := GetMonthlySummary(userID, monthYear)
//		if monthlySummary != nil {
//			return monthlySummary, nil
//		}
//		monthlySummary, err := CreateMonthlySummary(userID, monthYear)
//		if err != nil {
//			log.Printf("Failed to create monthly summary: %v", err)
//			return nil, fmt.Errorf("failed to create monthly summary: %v", err)
//		}
//		return monthlySummary, nil
//	}
func CreateMonthlySummary(userID int, monthYear int, totalSpent float64, startingBalance float64, income float64, savedAmount float64, invested float64, fixedExpenses float64, savingTargetPercentage float64, budget float64) (*MonthlySummary, error) {
	query := "INSERT INTO monthly_summary (user_id, monthyear, total_spent, starting_balance, income, saved_amount, invested, fixed_expenses, saving_target_percentage) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, user_id, monthyear, total_spent, starting_balance, income, saved_amount, invested, fixed_expenses, saving_target_percentage, created_at, updated_at"
	var monthlySummary MonthlySummary

	err := DB.QueryRow(query, userID, monthYear, totalSpent, startingBalance, income, savedAmount, invested, fixedExpenses, savingTargetPercentage).Scan(&monthlySummary.ID, &monthlySummary.UserID, &monthlySummary.MonthYear, &monthlySummary.TotalSpent, &monthlySummary.StartingBalance, &monthlySummary.Income, &monthlySummary.SavedAmount, &monthlySummary.Invested, &monthlySummary.FixedExpenses, &monthlySummary.SavingTargetPercentage, &monthlySummary.CreatedAt, &monthlySummary.UpdatedAt)
	if err != nil {
		log.Printf("Failed to create monthly summary: %v", err)
		return nil, fmt.Errorf("failed to create monthly summary: %v", err)
	}

	_, err = CreateMonthlyBudgetSpendCategory(userID, monthlySummary.ID, monthYear, "general", budget)
	if err != nil {
		log.Printf("Failed to create monthly budget spend category: %v", err)
		return nil, fmt.Errorf("failed to create monthly budget spend category: %v", err)
	}
	return &monthlySummary, nil
}

func HasAnyMonthlySummaries(userID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM monthly_summary WHERE user_id = $1"
	err := DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to count monthly summaries: %v", err)
	}
	return count > 0, nil
}

func GetMonthlySummary(userID int, monthYear int) (*MonthlySummary, error) {
	query := "SELECT id, user_id, monthyear, total_spent, starting_balance, income, saved_amount, invested, fixed_expenses, saving_target_percentage, created_at, updated_at FROM monthly_summary WHERE user_id = $1 AND monthyear = $2"
	var monthlySummary MonthlySummary
	err := DB.QueryRow(query, userID, monthYear).Scan(&monthlySummary.ID, &monthlySummary.UserID, &monthlySummary.MonthYear, &monthlySummary.TotalSpent, &monthlySummary.StartingBalance, &monthlySummary.Income, &monthlySummary.SavedAmount, &monthlySummary.Invested, &monthlySummary.FixedExpenses, &monthlySummary.SavingTargetPercentage, &monthlySummary.CreatedAt, &monthlySummary.UpdatedAt)
	if err != nil {
		log.Printf("Failed to get monthly summary: %v", err)
		return nil, fmt.Errorf("failed to get monthly summary: %v", err)
	}
	return &monthlySummary, nil
}

func UpdateMonthlySummaryTotalSpent(monthlySummary MonthlySummary) (*MonthlySummary, error) {
	query := "UPDATE monthly_summary SET total_spent = $1 WHERE id = $2 RETURNING id, user_id, monthyear, total_spent, starting_balance, income, saved_amount, invested, fixed_expenses, saving_target_percentage, created_at, updated_at"
	var updatedMonthlySummary MonthlySummary
	err := DB.QueryRow(query, monthlySummary.TotalSpent, monthlySummary.ID).Scan(&updatedMonthlySummary.ID, &updatedMonthlySummary.UserID, &updatedMonthlySummary.MonthYear, &updatedMonthlySummary.TotalSpent, &updatedMonthlySummary.StartingBalance, &updatedMonthlySummary.Income, &updatedMonthlySummary.SavedAmount, &updatedMonthlySummary.Invested, &updatedMonthlySummary.FixedExpenses, &updatedMonthlySummary.SavingTargetPercentage, &updatedMonthlySummary.CreatedAt, &updatedMonthlySummary.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update monthly summary: %v", err)
	}
	return &updatedMonthlySummary, nil
}

func UpdateMonthlySummary(userID int, monthYear int, totalSpent float64, startingBalance float64, income float64, savedAmount float64, invested float64, fixedExpenses float64, savingTargetPercentage float64) (*MonthlySummary, error) {
	query := "UPDATE monthly_summary SET total_spent = $1, starting_balance = $2, income = $3, saved_amount = $4, invested = $5, fixed_expenses = $6, saving_target_percentage = $7 WHERE user_id = $8 AND monthyear = $9 RETURNING id, user_id, monthyear, total_spent, starting_balance, income, saved_amount, invested, fixed_expenses, saving_target_percentage, created_at, updated_at"
	var monthlySummary MonthlySummary
	err := DB.QueryRow(query, totalSpent, startingBalance, income, savedAmount, invested, fixedExpenses, savingTargetPercentage, userID, monthYear).Scan(&monthlySummary.ID, &monthlySummary.UserID, &monthlySummary.MonthYear, &monthlySummary.TotalSpent, &monthlySummary.StartingBalance, &monthlySummary.Income, &monthlySummary.SavedAmount, &monthlySummary.Invested, &monthlySummary.FixedExpenses, &monthlySummary.SavingTargetPercentage, &monthlySummary.CreatedAt, &monthlySummary.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update monthly summary: %v", err)
	}
	return &monthlySummary, nil
}

// ********** MONTHLY BUDGET SPEND CATEGORY **********

func GetMonthlyBudgetSpendCategory(userID int, monthlySummaryID int, monthYear int, category string) (*MonthlyBudgetSpendCategory, error) {
	query := "SELECT id, user_id, monthly_summary_id, month_year, category, budget, total_spent, daily_allowance, created_at, updated_at FROM monthly_budget_spend_category WHERE user_id = $1 AND monthly_summary_id = $2 AND month_year = $3 AND category = $4"
	var monthlyBudgetSpendCategory MonthlyBudgetSpendCategory
	err := DB.QueryRow(query, userID, monthlySummaryID, monthYear, category).Scan(&monthlyBudgetSpendCategory.ID, &monthlyBudgetSpendCategory.UserID, &monthlyBudgetSpendCategory.MonthlySummaryID, &monthlyBudgetSpendCategory.MonthYear, &monthlyBudgetSpendCategory.Category, &monthlyBudgetSpendCategory.Budget, &monthlyBudgetSpendCategory.TotalSpent, &monthlyBudgetSpendCategory.DailyAllowance, &monthlyBudgetSpendCategory.CreatedAt, &monthlyBudgetSpendCategory.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly budget spend category: %v", err)
	}
	return &monthlyBudgetSpendCategory, nil
}

func CreateMonthlyBudgetSpendCategory(userID int, monthlySummaryID int, monthYear int, category string, budget float64) (*MonthlyBudgetSpendCategory, error) {
	query := "INSERT INTO monthly_budget_spend_category (user_id, monthly_summary_id, month_year, category, budget, total_spent) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_id, monthly_summary_id, month_year, category, budget, total_spent, created_at, updated_at"
	var monthlyBudgetSpendCategory MonthlyBudgetSpendCategory
	err := DB.QueryRow(query, userID, monthlySummaryID, monthYear, category, budget, 0).Scan(&monthlyBudgetSpendCategory.ID, &monthlyBudgetSpendCategory.UserID, &monthlyBudgetSpendCategory.MonthlySummaryID, &monthlyBudgetSpendCategory.MonthYear, &monthlyBudgetSpendCategory.Category, &monthlyBudgetSpendCategory.Budget, &monthlyBudgetSpendCategory.TotalSpent, &monthlyBudgetSpendCategory.CreatedAt, &monthlyBudgetSpendCategory.UpdatedAt)
	if err != nil {
		log.Printf("Failed to create monthly budget spend category: %v", err)
		return nil, fmt.Errorf("failed to create monthly budget spend category: %v", err)
	}
	return &monthlyBudgetSpendCategory, nil
}

func GetMonthlyBudgetSpendCategories(monthlySummaryID int) ([]MonthlyBudgetSpendCategory, float64, error) {
	query := "SELECT id, user_id, monthly_summary_id, month_year, category, budget, total_spent, daily_allowance, created_at, updated_at FROM monthly_budget_spend_category WHERE monthly_summary_id = $1"
	var monthlyBudgetSpendCategories []MonthlyBudgetSpendCategory
	rows, err := DB.Query(query, monthlySummaryID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get monthly budget spend categories: %v", err)
	}
	defer rows.Close()
	totalDailyAllowance := 0.0
	for rows.Next() {
		var monthlyBudgetSpendCategory MonthlyBudgetSpendCategory
		err := rows.Scan(&monthlyBudgetSpendCategory.ID, &monthlyBudgetSpendCategory.UserID, &monthlyBudgetSpendCategory.MonthlySummaryID, &monthlyBudgetSpendCategory.MonthYear, &monthlyBudgetSpendCategory.Category, &monthlyBudgetSpendCategory.Budget, &monthlyBudgetSpendCategory.TotalSpent, &monthlyBudgetSpendCategory.DailyAllowance, &monthlyBudgetSpendCategory.CreatedAt, &monthlyBudgetSpendCategory.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan monthly budget spend category: %v", err)
		}
		monthlyBudgetSpendCategories = append(monthlyBudgetSpendCategories, monthlyBudgetSpendCategory)
		totalDailyAllowance += monthlyBudgetSpendCategory.DailyAllowance
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating monthly budget spend categories: %v", err)
	}
	return monthlyBudgetSpendCategories, totalDailyAllowance, nil
}

func UpdateMonthlyBudgetSpendCategory(monthlyBudgetSpendCategory MonthlyBudgetSpendCategory) error {
	query := "UPDATE monthly_budget_spend_category SET total_spent = $1, daily_allowance = $2 WHERE id = $3"
	_, err := DB.Exec(query, monthlyBudgetSpendCategory.TotalSpent, monthlyBudgetSpendCategory.DailyAllowance, monthlyBudgetSpendCategory.ID)
	if err != nil {
		return fmt.Errorf("failed to update monthly budget spend category: %v", err)
	}
	return nil
}

// ********** MONTHLY BALANCE **********

func GetOrCreateMonthlyBalance(userID int, monthYear int) (*MonthlyBalance, error) {
	monthlyBalance, _ := GetMonthlyBalance(userID, monthYear)
	if monthlyBalance != nil {
		return monthlyBalance, nil
	}
	monthlyBalance, err := CreateMonthlyBalance(userID, monthYear)
	if err != nil {
		log.Printf("Failed to create monthly balance: %v", err)
		return nil, fmt.Errorf("failed to create monthly balance: %v", err)
	}
	return monthlyBalance, nil
}

func HasAnyMonthlyBalances(userID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM monthly_balance WHERE user_id = $1"
	err := DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to count monthly balances: %v", err)
	}
	return count > 0, nil
}

func GetMonthlyBalance(userID int, monthYear int) (*MonthlyBalance, error) {
	query := "SELECT id, user_id, monthyear, total_owing, net_cash, available_balance, current_balance, created_at, updated_at FROM monthly_balance WHERE user_id = $1 AND monthyear = $2"
	var monthlyBalance MonthlyBalance
	err := DB.QueryRow(query, userID, monthYear).Scan(&monthlyBalance.ID, &monthlyBalance.UserID, &monthlyBalance.MonthYear, &monthlyBalance.TotalOwing, &monthlyBalance.NetCash, &monthlyBalance.AvailableBalance, &monthlyBalance.CurrentBalance, &monthlyBalance.CreatedAt, &monthlyBalance.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly balance: %v", err)
	}
	return &monthlyBalance, nil
}

func CreateMonthlyBalance(userID int, monthYear int) (*MonthlyBalance, error) {
	query := "INSERT INTO monthly_balance (user_id, monthyear, total_owing, net_cash, available_balance, current_balance) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_id, monthyear, total_owing, net_cash, available_balance, current_balance, created_at, updated_at"
	var monthlyBalance MonthlyBalance
	err := DB.QueryRow(query, userID, monthYear, 0, 0, 0, 0).Scan(&monthlyBalance.ID, &monthlyBalance.UserID, &monthlyBalance.MonthYear, &monthlyBalance.TotalOwing, &monthlyBalance.NetCash, &monthlyBalance.AvailableBalance, &monthlyBalance.CurrentBalance, &monthlyBalance.CreatedAt, &monthlyBalance.UpdatedAt)
	if err != nil {
		log.Printf("Failed to create monthly balance: %v", err)
		return nil, fmt.Errorf("failed to create monthly balance: %v", err)
	}
	return &monthlyBalance, nil
}

func UpdateMonthlyBalance(userID int, monthYear int, totalOwing float64, netCash float64, availableBalance float64, currentBalance float64) (*MonthlyBalance, error) {
	query := "UPDATE monthly_balance SET total_owing = $1, net_cash = $2, available_balance = $3, current_balance = $4 WHERE user_id = $5 AND monthyear = $6 RETURNING id, user_id, monthyear, total_owing, net_cash, available_balance, current_balance, created_at, updated_at"
	var monthlyBalance MonthlyBalance
	err := DB.QueryRow(query, totalOwing, netCash, availableBalance, currentBalance, userID, monthYear).Scan(&monthlyBalance.ID, &monthlyBalance.UserID, &monthlyBalance.MonthYear, &monthlyBalance.TotalOwing, &monthlyBalance.NetCash, &monthlyBalance.AvailableBalance, &monthlyBalance.CurrentBalance, &monthlyBalance.CreatedAt, &monthlyBalance.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update monthly balance: %v", err)
	}
	return &monthlyBalance, nil
}

// ********** SAVING GOALS **********

func GetSavingsGoals(userID int) ([]SavingsGoal, error) {
	query := "SELECT id, user_id, name, currently_saved, total, redeemed, created_at, updated_at FROM saving_goal WHERE user_id = $1"
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query savings goals: %v", err)
	}
	defer rows.Close()

	var savingsGoals []SavingsGoal
	for rows.Next() {
		var savingsGoal SavingsGoal
		err := rows.Scan(&savingsGoal.ID, &savingsGoal.UserID, &savingsGoal.Name, &savingsGoal.CurrentSaved, &savingsGoal.TotalAmount, &savingsGoal.Redeemed, &savingsGoal.CreatedAt, &savingsGoal.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan savings goal: %v", err)
		}
		savingsGoals = append(savingsGoals, savingsGoal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating savings goals: %v", err)
	}

	return savingsGoals, nil
}

func CreateSavingsGoal(userID int, name string, totalAmount float64, currentSaved float64) (*SavingsGoal, error) {
	query := "INSERT INTO saving_goal (user_id, name, total, redeemed, currently_saved) VALUES ($1, $2, $3, $4, $5) RETURNING id, user_id, name, total, redeemed, currently_saved, created_at, updated_at"
	var savingsGoal SavingsGoal
	err := DB.QueryRow(query, userID, name, totalAmount, false, currentSaved).Scan(&savingsGoal.ID, &savingsGoal.UserID, &savingsGoal.Name, &savingsGoal.TotalAmount, &savingsGoal.Redeemed, &savingsGoal.CurrentSaved, &savingsGoal.CreatedAt, &savingsGoal.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create savings goal: %v", err)
	}
	return &savingsGoal, nil
}
