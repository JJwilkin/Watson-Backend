package plaid

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"watson/database"

	"github.com/joho/godotenv"
	plaid "github.com/plaid/plaid-go/v31/plaid"
)

var (
	PLAID_CLIENT_ID                      = ""
	PLAID_SECRET                         = ""
	PLAID_ENV                            = "production"
	PLAID_PRODUCTS                       = "transactions"
	PLAID_COUNTRY_CODES                  = []string{"US", "CA"}
	PLAID_REDIRECT_URI                   = ""
	APP_PORT                             = ""
	Client              *plaid.APIClient = nil
)

func InitPlaid() {
	// load env vars from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error when loading environment variables from .env file %w", err)
	}

	// set constants from env
	PLAID_CLIENT_ID = os.Getenv("PLAID_CLIENT_ID")
	PLAID_SECRET = os.Getenv("PLAID_SECRET")

	if PLAID_CLIENT_ID == "" || PLAID_SECRET == "" {
		log.Fatal("Error: PLAID_SECRET or PLAID_CLIENT_ID is not set. Did you copy .env.example to .env and fill it out?")
	}

	PLAID_ENV = os.Getenv("PLAID_ENV")
	PLAID_PRODUCTS = os.Getenv("PLAID_PRODUCTS")

	// PLAID_REDIRECT_URI = os.Getenv("PLAID_REDIRECT_URI")
	APP_PORT = os.Getenv("APP_PORT")

	// set defaults
	if PLAID_PRODUCTS == "" {
		PLAID_PRODUCTS = "transactions"
	}
	if PLAID_ENV == "" {
		PLAID_ENV = "sandbox"
	}
	if APP_PORT == "" {
		APP_PORT = "8000"
	}
	if PLAID_CLIENT_ID == "" {
		log.Fatal("PLAID_CLIENT_ID is not set. Make sure to fill out the .env file")
	}
	if PLAID_SECRET == "" {
		log.Fatal("PLAID_SECRET is not set. Make sure to fill out the .env file")
	}

	// create Plaid client
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", PLAID_CLIENT_ID)
	configuration.AddDefaultHeader("PLAID-SECRET", PLAID_SECRET)
	// Set the correct environment using proper Plaid constants
	if PLAID_ENV == "production" {
		configuration.UseEnvironment(plaid.Production)
	} else if PLAID_ENV == "sandbox" {
		configuration.UseEnvironment(plaid.Sandbox)
	} else {
		log.Fatal("Invalid PLAID_ENV. Must be either 'production' or 'sandbox'")
	}
	Client = plaid.NewAPIClient(configuration)
	log.Printf("Plaid client initialized")
}

func CreateLinkToken(userIdInt int) (string, error) {

	request := plaid.NewLinkTokenCreateRequest(
		"Watson",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_CA},
		*plaid.NewLinkTokenCreateRequestUser(strconv.Itoa(userIdInt)),
	)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})
	// request.SetWebhook("https://sample-web-hook.com")
	// request.SetRedirectUri("https://domainname.com/oauth-page.html")

	linkTokenCreateResp, _, err := Client.PlaidApi.LinkTokenCreate(context.Background()).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		log.Printf("Failed to create link token: %v", err)
		return "", err
	}
	linkToken := linkTokenCreateResp.GetLinkToken()

	return linkToken, nil
}

func ExchangePublicToken(publicToken string, userIdInt int) (string, string, error) {
	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	exchangePublicTokenResp, _, err := Client.PlaidApi.ItemPublicTokenExchange(context.Background()).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	accessToken := exchangePublicTokenResp.GetAccessToken()
	itemId := exchangePublicTokenResp.GetItemId()
	log.Printf("Plaid success payload: %v, %v", itemId, accessToken)
	err = database.CreatePlaidToken(userIdInt, accessToken, itemId)
	if err != nil {
		log.Printf("Failed to create plaid token: %v", err)
		return "", "", err
	}
	return accessToken, itemId, nil
}

func GetTransactions(accessToken string, startDate string, endDate string) ([]plaid.Transaction, error) {
	const iso8601TimeFormat = "2006-01-02"

	startDate = time.Now().Add(-365 * 24 * time.Hour).Format(iso8601TimeFormat)
	endDate = time.Now().Format(iso8601TimeFormat)
	// options := plaid.TransactionsGetRequestOptions{
	// 	IncludePersonalFinanceCategory := true,
	// }
	// request.SetOptions(options)

	request := plaid.NewTransactionsGetRequest(
		accessToken,
		startDate,
		endDate,
	)

	// options := plaid.TransactionsGetRequestOptions{
	// Count:  plaid.PtrInt32(100),
	// Offset: plaid.PtrInt32(0),
	// IncludePersonalFinanceCategory: true,
	// }

	// request.SetOptions(options)

	transactionsResp, _, err := Client.PlaidApi.TransactionsGet(context.Background()).TransactionsGetRequest(*request).Execute()
	if err != nil {
		log.Printf("Failed to get transactions: %v", err)
		return nil, err
	}
	return transactionsResp.GetTransactions(), nil
}

func GetAccounts(accessToken string) ([]plaid.AccountBase, error) {
	// options := plaid.TransactionsGetRequestOptions{
	// 	IncludePersonalFinanceCategory := true,
	// }
	// request.SetOptions(options)

	request := plaid.NewAccountsGetRequest(
		accessToken,
	)

	// options := plaid.TransactionsGetRequestOptions{
	// Count:  plaid.PtrInt32(100),
	// Offset: plaid.PtrInt32(0),
	// IncludePersonalFinanceCategory: true,
	// }

	// request.SetOptions(options)

	accountsResp, _, err := Client.PlaidApi.AccountsGet(context.Background()).AccountsGetRequest(*request).Execute()
	if err != nil {
		log.Printf("Failed to get accounts: %v", err)
		return nil, err
	}
	accounts := accountsResp.GetAccounts()

	for _, account := range accounts {
		log.Printf("Account: %v", account)
	}

	return accounts, nil
}
