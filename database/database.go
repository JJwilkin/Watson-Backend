package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

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
	TransactionID   int       `json:"transaction_id"`
	UserID          int       `json:"user_id"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	TransactionDate time.Time `json:"transaction_date"`
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

// Monthly Summary
type BudgetCategory struct {
	Category      string                   `json:"category"`
	Amount        float64                  `json:"amount"`
	SubCategories []map[string]interface{} `json:"sub_categories"`
}

type MonthlySummary struct {
	ID                     int             `json:"id"`
	UserID                 int             `json:"user_id"`
	MonthYear              int             `json:"monthyear"`
	TotalSpent             float64         `json:"total_spent"`
	Budget                 json.RawMessage `json:"budget"`
	FixedExpenses          float64         `json:"fixed_expenses"`
	SavingTargetPercentage float64         `json:"saving_target_percentage"`
	StartingBalance        float64         `json:"starting_balance"`
	Income                 float64         `json:"income"`
	SavedAmount            float64         `json:"saved_amount"`
	Invested               float64         `json:"invested"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
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

		// Handle empty category array
		var category string
		if categories := transaction.GetCategory(); len(categories) > 0 {
			category = categories[0]
		} else {
			category = ""
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

func GetOrCreateMonthlySummary(userID int, monthYear int) (*MonthlySummary, error) {
	monthlySummary, _ := GetMonthlySummary(userID, monthYear)
	if monthlySummary != nil {
		return monthlySummary, nil
	}
	monthlySummary, err := CreateMonthlySummary(userID, monthYear)
	if err != nil {
		log.Printf("Failed to create monthly summary: %v", err)
		return nil, fmt.Errorf("failed to create monthly summary: %v", err)
	}
	return monthlySummary, nil
}
func CreateMonthlySummary(userID int, monthYear int) (*MonthlySummary, error) {
	query := "INSERT INTO monthly_summary (user_id, monthyear, total_spent, budget, starting_balance, income, saved_amount, invested, fixed_expenses, saving_target_percentage) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, user_id, monthyear, total_spent, budget, starting_balance, income, saved_amount, invested, fixed_expenses, saving_target_percentage, created_at, updated_at"
	var monthlySummary MonthlySummary
	defaultBudget := []BudgetCategory{
		{
			Category:      "general",
			Amount:        0.0,
			SubCategories: []map[string]interface{}{},
		},
	}
	budgetJSON, _ := json.Marshal(defaultBudget)
	err := DB.QueryRow(query, userID, monthYear, 0.0, budgetJSON, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0).Scan(&monthlySummary.ID, &monthlySummary.UserID, &monthlySummary.MonthYear, &monthlySummary.TotalSpent, &monthlySummary.Budget, &monthlySummary.StartingBalance, &monthlySummary.Income, &monthlySummary.SavedAmount, &monthlySummary.Invested, &monthlySummary.FixedExpenses, &monthlySummary.SavingTargetPercentage, &monthlySummary.CreatedAt, &monthlySummary.UpdatedAt)
	if err != nil {
		log.Printf("Failed to create monthly summary: %v", err)
		return nil, fmt.Errorf("failed to create monthly summary: %v", err)
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
	query := "SELECT id, user_id, monthyear, total_spent, budget, starting_balance, income, saved_amount, invested, fixed_expenses, saving_target_percentage, created_at, updated_at FROM monthly_summary WHERE user_id = $1 AND monthyear = $2"
	var monthlySummary MonthlySummary
	err := DB.QueryRow(query, userID, monthYear).Scan(&monthlySummary.ID, &monthlySummary.UserID, &monthlySummary.MonthYear, &monthlySummary.TotalSpent, &monthlySummary.Budget, &monthlySummary.StartingBalance, &monthlySummary.Income, &monthlySummary.SavedAmount, &monthlySummary.Invested, &monthlySummary.FixedExpenses, &monthlySummary.SavingTargetPercentage, &monthlySummary.CreatedAt, &monthlySummary.UpdatedAt)
	if err != nil {
		log.Printf("Failed to get monthly summary: %v", err)
		return nil, fmt.Errorf("failed to get monthly summary: %v", err)
	}
	return &monthlySummary, nil
}

func UpdateMonthlySummary(userID int, monthYear int, totalSpent float64, budget json.RawMessage, startingBalance float64, income float64, savedAmount float64, invested float64, fixedExpenses float64, savingTargetPercentage float64) (*MonthlySummary, error) {
	query := "UPDATE monthly_summary SET total_spent = $1, budget = $2, starting_balance = $3, income = $4, saved_amount = $5, invested = $6, fixed_expenses = $7, saving_target_percentage = $8 WHERE user_id = $9 AND monthyear = $10 RETURNING id, user_id, monthyear, total_spent, budget, starting_balance, income, saved_amount, invested, fixed_expenses, saving_target_percentage, created_at, updated_at"
	var monthlySummary MonthlySummary
	err := DB.QueryRow(query, totalSpent, budget, startingBalance, income, savedAmount, invested, fixedExpenses, savingTargetPercentage, userID, monthYear).Scan(&monthlySummary.ID, &monthlySummary.UserID, &monthlySummary.MonthYear, &monthlySummary.TotalSpent, &monthlySummary.Budget, &monthlySummary.StartingBalance, &monthlySummary.Income, &monthlySummary.SavedAmount, &monthlySummary.Invested, &monthlySummary.FixedExpenses, &monthlySummary.SavingTargetPercentage, &monthlySummary.CreatedAt, &monthlySummary.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update monthly summary: %v", err)
	}
	return &monthlySummary, nil
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

// GetBudgetAsStruct converts the raw JSON budget to structured BudgetCategory
func (ms *MonthlySummary) GetBudgetAsStruct() ([]BudgetCategory, error) {
	if ms.Budget == nil {
		return []BudgetCategory{}, nil
	}

	var budgetCategories []BudgetCategory
	err := json.Unmarshal(ms.Budget, &budgetCategories)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal budget: %v", err)
	}
	return budgetCategories, nil
}

// SetBudgetFromStruct converts structured BudgetCategory to raw JSON
func (ms *MonthlySummary) SetBudgetFromStruct(budgetCategories []BudgetCategory) error {
	budgetJSON, err := json.Marshal(budgetCategories)
	if err != nil {
		return fmt.Errorf("failed to marshal budget: %v", err)
	}
	ms.Budget = budgetJSON
	return nil
}
