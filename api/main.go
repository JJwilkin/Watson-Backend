package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"watson/database"

	plaid "watson/plaid"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	// plaid "github.com/plaid/plaid-go/v31/plaid"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// User represents a user registration request
type User struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=2"`
}

// TransactionRequest represents a transaction creation request
type TransactionRequest struct {
	Description     string    `json:"description" binding:"required"`
	Amount          float64   `json:"amount" binding:"required"`
	TransactionDate time.Time `json:"transaction_date" binding:"required"`
}

// AUTH

func login(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	dbUser, err := database.GetUserByEmailAndPassword(user.Email, user.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	jwt, err := GenerateJWTWithDefaultExpiry(dbUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate JWT",
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"user_id": dbUser.UserID,
			"email":   dbUser.Email,
		},
		"jwt": jwt,
	})
}

func register(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create user in database
	dbUser, err := database.CreateUser(user.Email, user.Password)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register user",
		})
		return
	}
	jwt, err := GenerateJWTWithDefaultExpiry(dbUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate JWT",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"user_id": dbUser.UserID,
			"email":   dbUser.Email,
		},
		"jwt": jwt,
	})
}

func getBalance(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}

	c.JSON(http.StatusOK, gin.H{
		"balances": gin.H{
			"month_remaining": 423,
			"daily_remaining": 12,
		},
		"user_id": userIdInt,
	})
}

// Get user by ID
func getUser(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}

	user, err := database.GetUserByID(userIdInt)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"user_id": user.UserID,
			"email":   user.Email,
		},
	})
}

func genereateBankLink(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	bankLinkUrl := GetBankLinkURL()
	temporaryJWT, err := GenerateTemporaryJWT(userIdInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate temporary JWT",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"bank_link_url": bankLinkUrl + "?jwt=" + temporaryJWT,
	})
}

// func generatePlaidLink(c *gin.Context) {
// 	userIdInt, err := AuthMiddleware(c)
//     if err != nil {
//         return // AuthMiddleware already sent the response
//     }
//     plaidLinkUrl := GetPlaidLinkURL()
// }

// func createLinkToken(c *gin.Context) {
//     ctx := context.Background()

//     // Get the client_user_id by searching for the current user
//     user, _ := usermodels.Find(...)
//     clientUserId := user.ID.String()

//     // Create a link_token for the given user
//     request := plaid.NewLinkTokenCreateRequest("Plaid Test App", "en", []plaid.CountryCode{plaid.COUNTRYCODE_US}, *plaid.NewLinkTokenCreateRequestUser(clientUserId))
//     request.SetWebhook("https://webhook.sample.com")
//     request.SetRedirectUri("https://domainname.com/oauth-page.html")
//     request.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH})

//     resp, _, err := testClient.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()

//     // Send the data to the client
//     c.JSON(http.StatusOK, gin.H{
//       "link_token": resp.GetLinkToken(),
//     })
//   }

func handleTellerSuccess(c *gin.Context) {
	var payload database.TellerPayload
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	log.Printf("Teller success payload: %v", payload)
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	tellerInstitution, err := database.HandleTellerSuccess(userIdInt, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to handle teller success",
		})
		return
	}

	// Enqueue job to process transactions
	jobData := map[string]interface{}{
		"user_id": userIdInt,
		"token":   payload.AccessToken,
	}

	// jobData := map[string]interface{}{
	// 	"user_id":            userIdInt,
	// 	"teller_institution": tellerInstitution,
	// 	"payload":            payload,
	// }
	log.Printf("Job data: %v", jobData)

	// Send POST request to background worker to enqueue job
	enqueueRequest := map[string]interface{}{
		"type": "new_teller_link",
		"data": jobData,
	}

	enqueueJSON, err := json.Marshal(enqueueRequest)
	if err != nil {
		log.Printf("Failed to marshal enqueue request: %v", err)
	} else {
		// Make HTTP request to background worker
		resp, err := http.Post("http://localhost:8081/enqueue", "application/json", bytes.NewBuffer(enqueueJSON))
		// resp, err := http.Post("http://worker:8081/enqueue", "application/json", bytes.NewBuffer(enqueueJSON))
		if err != nil {
			log.Printf("Failed to enqueue job: %v", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				log.Printf("Successfully enqueued transaction processing job for user %d", userIdInt)
			} else {
				log.Printf("Failed to enqueue job, status: %d", resp.StatusCode)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Teller success handled successfully",
		"teller_institution": tellerInstitution,
	})
}

// Health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

func createLinkToken(c *gin.Context) {
	// Get the client_user_id by searching for the current user
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}

	linkToken, err := plaid.CreateLinkToken(userIdInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create link token",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"link_token": linkToken,
	})
}

func handlePlaidSuccess(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	accessToken, itemId, err := plaid.ExchangePublicToken(payload["public_token"].(string), userIdInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to exchange public token",
		})
		return
	}

	jobData := map[string]interface{}{
		"user_id":      userIdInt,
		"access_token": accessToken,
		"item_id":      itemId,
	}

	enqueueRequest := map[string]interface{}{
		"type": "initial_plaid_sync",
		"data": jobData,
	}

	enqueueJSON, err := json.Marshal(enqueueRequest)
	if err != nil {
		log.Printf("Failed to marshal enqueue request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to marshal enqueue request",
		})
		return
	}

	// Make HTTP request to background worker
	resp, err := http.Post("http://localhost:8081/enqueue", "application/json", bytes.NewBuffer(enqueueJSON))
	// resp, err := http.Post("http://worker:8081/enqueue", "application/json", bytes.NewBuffer(enqueueJSON))
	if err != nil {
		log.Printf("Failed to enqueue job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to enqueue job",
		})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully enqueued transaction processing job for user %d", userIdInt)
	} else {
		log.Printf("Failed to enqueue job, status: %d", resp.StatusCode)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to enqueue job",
		})
		return
	}

	// Send response to client
	c.JSON(http.StatusOK, gin.H{
		"message": "Plaid success handled successfully",
	})

}

func getPlaidTransactions(c *gin.Context) {
	_, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	accessToken := c.Query("access_token")

	transactions, err := plaid.GetTransactions(accessToken, "2025-01-01", "2025-01-31")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get transactions",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
	})
}

func getPlaidAccounts(c *gin.Context) {
	_, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	accessToken := c.Query("access_token")
	accounts, err := plaid.GetAccounts(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get accounts",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

// ** MONTHLY SUMMARY **

func getOrCreateMonthlySummary(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	monthYear := GetCurrentMonthYear()
	monthlySummary, err := database.GetOrCreateMonthlySummary(userIdInt, monthYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get or create monthly summary",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"monthly_summary": monthlySummary,
	})
}

func updateMonthlySummary(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	monthYear := GetCurrentMonthYear()
	if monthYearFromPayload, exists := payload["month_year"]; exists {
		monthYear = int(monthYearFromPayload.(float64))
	}

	monthlySummary, err := database.GetOrCreateMonthlySummary(userIdInt, monthYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get or create monthly summary",
		})
		return
	}

	// Update only the fields that are provided in the payload
	totalSpent := monthlySummary.TotalSpent
	if val, exists := payload["total_spent"]; exists {
		totalSpent = val.(float64)
	}

	budget := monthlySummary.Budget
	if val, exists := payload["budget"]; exists {
		budgetJSON, _ := json.Marshal(val)
		budget = budgetJSON
	}

	startingBalance := monthlySummary.StartingBalance
	if val, exists := payload["starting_balance"]; exists {
		startingBalance = val.(float64)
	}

	income := monthlySummary.Income
	if val, exists := payload["income"]; exists {
		income = val.(float64)
	}

	savedAmount := monthlySummary.SavedAmount
	if val, exists := payload["saved_amount"]; exists {
		savedAmount = val.(float64)
	}

	invested := monthlySummary.Invested
	if val, exists := payload["invested"]; exists {
		invested = val.(float64)
	}

	monthlySummary, err = database.UpdateMonthlySummary(userIdInt, monthYear, totalSpent, budget, startingBalance, income, savedAmount, invested)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update monthly summary",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"monthly_summary": monthlySummary,
	})
}

// ** MONTHLY BALANCE **

func getOrCreateMonthlyBalance(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	monthYear := GetCurrentMonthYear()
	if val, exists := c.GetQuery("monthyear"); exists {
		if parsed, err := strconv.Atoi(val); err == nil {
			monthYear = parsed
		}
	}
	monthlyBalance, err := database.GetOrCreateMonthlyBalance(userIdInt, monthYear)
	if err != nil {
		log.Printf("Failed to get or create monthly balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get or create monthly balance",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"monthly_balance": monthlyBalance,
	})
}

func updateMonthlyBalance(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	monthYear := GetCurrentMonthYear()
	if monthYearFromPayload, exists := payload["monthyear"]; exists {
		log.Printf("Month year from payload: %v", monthYearFromPayload)
		monthYear = int(monthYearFromPayload.(float64))
	}

	monthlyBalance, err := database.GetOrCreateMonthlyBalance(userIdInt, monthYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get or create monthly balance",
		})
		return
	}
	totalOwing := monthlyBalance.TotalOwing
	if val, exists := payload["total_owing"]; exists {
		totalOwing = val.(float64)
	}

	netCash := monthlyBalance.NetCash
	if val, exists := payload["net_cash"]; exists {
		netCash = val.(float64)
	}

	availableBalance := monthlyBalance.AvailableBalance
	if val, exists := payload["available_balance"]; exists {
		availableBalance = val.(float64)
	}

	currentBalance := monthlyBalance.CurrentBalance
	if val, exists := payload["current_balance"]; exists {
		currentBalance = val.(float64)
	}

	monthlyBalance, err = database.UpdateMonthlyBalance(userIdInt, monthYear, totalOwing, netCash, availableBalance, currentBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update monthly balance",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"monthly_balance": monthlyBalance,
	})
}

// ** SAVING GOALS **

func getSavingGoals(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	savingsGoals, err := database.GetSavingsGoals(userIdInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get savings goals",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"savings_goals": savingsGoals,
	})
}

func createSavingGoal(c *gin.Context) {

	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	name := payload["name"].(string)
	totalAmount := payload["total"].(float64)
	currentSaved := 0.0

	savingsGoal, err := database.CreateSavingsGoal(userIdInt, name, totalAmount, currentSaved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create savings goal",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"savings_goal": savingsGoal,
	})
}

func validateJWT(c *gin.Context) {
	userIdInt, err := AuthMiddleware(c)
	if err != nil {
		return // AuthMiddleware already sent the response
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id": userIdInt,
	})
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables or defaults")
	}

	// Get database connection string from environment variable
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "postgres://postgres:password@localhost:5432/watson?sslmode=disable" // fallback
	}

	plaid.InitPlaid()
	// Initialize shared database connection
	if err := database.InitDB(dbConnStr); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	router := gin.Default()

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // Must be false when AllowOrigins is "*"
	}))

	router.GET("/validate-jwt", validateJWT)
	// Routes
	router.POST("/register", register)
	router.GET("/users/", getUser)
	router.GET("/balances", getBalance)
	router.POST("/login", login)
	// Bank
	router.GET("/bank-link", genereateBankLink)
	router.POST("/bank-link-teller/success", handleTellerSuccess)
	router.GET("/create-link-token", createLinkToken)
	router.POST("/bank-link-plaid/success", handlePlaidSuccess)
	router.GET("/plaid/transactions", getPlaidTransactions)
	router.GET("/plaid/accounts", getPlaidAccounts)
	// Monthly Summary
	router.GET("/monthly-summary", getOrCreateMonthlySummary)
	router.PUT("/monthly-summary", updateMonthlySummary)
	// Monthly Balance
	router.GET("/monthly-balance", getOrCreateMonthlyBalance)
	router.PUT("/monthly-balance", updateMonthlyBalance)

	//Saving Goals
	router.GET("/saving-goals", getSavingGoals)
	router.POST("/saving-goal", createSavingGoal)
	// Health check
	router.GET("/health", healthCheck)

	config := LoadConfig()
	serverAddr := "0.0.0.0:" + config.ServerPort
	log.Printf("Server starting on %s", serverAddr)
	router.Run(serverAddr)
}
