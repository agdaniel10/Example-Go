package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// ------------------------------------------------------------------
// STRUCTS
// ------------------------------------------------------------------

type TransactionResult struct {
	Success bool
	Message string
	Balance float64
}

type Transaction struct {
	CustomerID      int
	TransactionType string
	Amount          float64
	Response        chan TransactionResult
}

type Account struct {
	CustomerId int
	Name       string
	Balance    float64
}

// ------------------------------------------------------------------
// SHARED STATE
// ------------------------------------------------------------------

var accounts = map[int]Account{
	1: {CustomerId: 1, Name: "Alice", Balance: 1000},
	2: {CustomerId: 2, Name: "Bob", Balance: 500},
	3: {CustomerId: 3, Name: "Charlie", Balance: 750},
}

// ------------------------------------------------------------------
// ATOMIC COUNTER
// ------------------------------------------------------------------

var totalTransactions = 0
var transactionLock sync.Mutex

func incrementTransactionCount() {
	transactionLock.Lock()
	defer transactionLock.Unlock()
	totalTransactions++
}

func getTransactionCount() int {
	transactionLock.Lock()
	defer transactionLock.Unlock()
	return totalTransactions
}

// ------------------------------------------------------------------
// ATOMIC COUNTER
// ------------------------------------------------------------------
var (
	transactionChannel = make(chan Transaction, 10)
	logChannel         = make(chan string)
	done               = make(chan bool)
)

// ------------------------------------------------------------------
// RATE LIMITER
// ------------------------------------------------------------------
var rateLimiterChannel = make(chan time.Time, 3)

func init() {
	// Prefill 3 tokens
	for range 3 {
		rateLimiterChannel <- time.Now()
	}
}

func refillRateLimiter() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			select {
			case rateLimiterChannel <- time.Now():
			default:
			}
		}
	}
}

// ------------------------------------------------------------------
// LOGGER GOROUTINE
// ------------------------------------------------------------------
func loggerWorker() {

	for {
		select {
		case <-done:
			return
		case msg := <-logChannel:
			fmt.Println(" [LOG]", msg)
		}
	}
}

func log(msg string) {
	logChannel <- msg
}

// ------------------------------------------------------------------
// STATE MANAGER GOROUTINE
// ------------------------------------------------------------------

func stateManager() {

	for {
		select {
		case <-done:
			return

		case tx := <-transactionChannel:
			account, exists := accounts[tx.CustomerID]
			if !exists {
				tx.Response <- TransactionResult{
					Success: false,
					Message: "Account not found",
					Balance: 0,
				}
				continue
			}

			// handle deposit
			if tx.TransactionType == "deposit" {
				account.Balance += tx.Amount
				accounts[tx.CustomerID] = account
				log(fmt.Sprintf("Deposit $%.2f → %s | New balance: $%.2f",
					tx.Amount, account.Name, account.Balance))
				tx.Response <- TransactionResult{
					Success: true,
					Message: "Deposit successful",
					Balance: account.Balance,
				}
			} else if tx.TransactionType == "withdraw" {
				if account.Balance < tx.Amount {
					log(fmt.Sprintf("Withdraw $%.2f ← %s | Insufficient funds!",
						tx.Amount, account.Name))
					tx.Response <- TransactionResult{
						Success: false,
						Message: "Insufficient Funds",
						Balance: account.Balance,
					}
				} else {
					account.Balance -= tx.Amount
					accounts[tx.CustomerID] = account
					log(fmt.Sprintf("Withdraw $%.2f ← %s | New balance: $%.2f",
						tx.Amount, account.Name, account.Balance))
					tx.Response <- TransactionResult{
						Success: true,
						Message: "Withdrawal successful",
						Balance: account.Balance,
					}
				}
			}
			incrementTransactionCount()
		}
	}
}

// ------------------------------------------------------------------
// TELLER WORKER POOL
// ------------------------------------------------------------------
func teller(tellerId int, jobs <-chan Transaction, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		select {
		case <-rateLimiterChannel:
		case <-time.After(2 * time.Second):
			log(fmt.Sprintf("Teller %d Rate limit hit! Skipping transaction.", tellerId))
			continue
		}

		tx := Transaction{
			CustomerID:      job.CustomerID,
			TransactionType: job.TransactionType,
			Amount:          job.Amount,
			Response:        make(chan TransactionResult, 1),
		}

		transactionChannel <- tx

		select {
		case result := <-tx.Response:
			log(fmt.Sprintf("Teller %d processed: %s", tellerId, result.Message))
		case <-time.After(3 * time.Second):
			log(fmt.Sprintf("Teller %d Transaction timed out!", tellerId))
		}

		time.Sleep(time.Duration(rand.Intn(200)+100) * time.Millisecond)
	}
}

// ------------------------------------------------------------------
// INTEREST TICKER
// ------------------------------------------------------------------
func interestTicker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return

		case <-ticker.C:
			for id, account := range accounts {
				interest := math.Round(account.Balance*0.01*100) / 100
				account.Balance += interest
				accounts[id] = account
			}

			log(" 💰 Interest applied to all accounts (1%) ")
		}
	}
}

func runWorkerPool() {
	fmt.Println("\n  🏃 Spawning 3 tellers to process 10 transactions...")

	jobs := make(chan Transaction, 10)
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go teller(i, jobs, &wg)
	}

	// Generate 10 random transactions
	for range 10 {
		jobs <- Transaction{
			CustomerID:      rand.Intn(3) + 1,
			TransactionType: []string{"deposit", "withdraw"}[rand.Intn(2)],
			Amount:          float64(rand.Intn(250) + 50),
		}
	}
	close(jobs)

	wg.Wait()
	fmt.Println("\n  All tellers done!")
}

// ------------------------------------------------------------------
// DISPLAY BALANCE
// ------------------------------------------------------------------

func displayBalance(customerID int) {

	fmt.Println(" Fetching balance for customer ", customerID, "...")
	defer fmt.Println(" Balance Check complete")

	account, ok := accounts[customerID]

	if !ok {
		fmt.Println(" Account not found")
		return

	}

	fmt.Println(" " + strings.Repeat("-", 30))
	fmt.Println(" Customer: ", account.Name)
	fmt.Println(" Balance : $", account.Balance)
	fmt.Println(" " + strings.Repeat("-", 30))

}

// ------------------------------------------------------------------
// MAIN GAME LOOP
// ------------------------------------------------------------------

func printHeader() {
	fmt.Println(" " + strings.Repeat("=", 45))
	fmt.Println("     🏦  BANK SIMULATION CLI GAME")
	fmt.Println(" " + strings.Repeat("=", 45))
}

func printMenu() {
	fmt.Println("\n  1. Deposit money")
	fmt.Println("  2. Withdraw money")
	fmt.Println("  3. Check balance")
	fmt.Println("  4. Run automatic transactions (worker pool)")
	fmt.Println("  5. View transaction count")
	fmt.Println("  6. Exit")
	fmt.Println()
}

func main() {

	go loggerWorker()
	go stateManager()
	go interestTicker()
	go refillRateLimiter()

	printHeader()

	for {
		printMenu()

		var choice string
		fmt.Println("  Enter your choice: ")
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			var cid int
			var amount float64

			fmt.Println(" Customer ID (1-3): ")
			_, err := fmt.Scanln(&cid)
			if err != nil {
				fmt.Println(" Invalid input. ")
				continue
			}

			fmt.Println(" Amount to deposit: $")
			_, err = fmt.Scanln(&amount)
			if err != nil {
				fmt.Println(" Invalid input")
				continue
			}

			// Build the transaction and send the state MANAGER
			tx := Transaction{
				CustomerID:      cid,
				TransactionType: "deposit",
				Amount:          amount,
				Response:        make(chan TransactionResult, 1),
			}

			transactionChannel <- tx

			select {
			case result := <-tx.Response:
				fmt.Printf("\n  %s | Balance: $%.2f\n", result.Message, result.Balance)
			case <-time.After(3 * time.Second):
				fmt.Println("   Transaction timed out.")
			}

		case "2":
			var cid int
			var amount float64

			fmt.Println(" Customer ID (1-3): ")
			_, err := fmt.Scanln(&cid)
			if err != nil {
				fmt.Println(" Invalid input. ")
				continue
			}

			fmt.Println(" Amount to deposit: $")
			_, err = fmt.Scanln(&amount)
			if err != nil {
				fmt.Println(" Invalid input")
				continue
			}

			// Build the transaction and send the state MANAGER
			tx := Transaction{
				CustomerID:      cid,
				TransactionType: "withdraw",
				Amount:          amount,
				Response:        make(chan TransactionResult, 1),
			}

			transactionChannel <- tx

			select {
			case result := <-tx.Response:
				fmt.Printf("\n  %s | Balance: $%.2f\n", result.Message, result.Balance)
			case <-time.After(3 * time.Second):
				fmt.Println("  ⚠️  Transaction timed out.")
			}

		case "3":
			var cid int
			fmt.Print(" Customer ID (1-3): ")
			_, err := fmt.Scanln(&cid)
			if err != nil {
				fmt.Println(" Invalid input. ")
				continue
			}

			displayBalance(cid)

		case "4":
			runWorkerPool()

		case "5":
			fmt.Printf("\n  Total transactions processed: %d\n\n", getTransactionCount())

		case "6":
			done <- true
			fmt.Println(" Thanks for banking with us! Goodbye ")
			return

		default:
			fmt.Println(" Invalid choice")
		}

		time.Sleep(300 * time.Millisecond)

	}
}
