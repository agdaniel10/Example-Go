# 🏦 Bank Simulation CLI — Go

A concurrent bank simulation game built in Go, demonstrating real-world concurrency patterns including goroutines, channels, worker pools, rate limiting, tickers, and mutexes.

> This project was built as a learning exercise to port a Python CLI simulation into Go, applying concepts from [Go by Example](https://gobyexample.com).

---

## 📋 Features

- Deposit and withdraw money from customer accounts
- Real-time transaction logging via a dedicated logger goroutine
- Automatic interest accrual every 5 seconds via a ticker
- Worker pool of 3 concurrent tellers processing transactions
- Bursty rate limiting on transactions
- Atomic transaction counter with mutex protection
- Stateful goroutine that safely owns and manages all account state

---

## 🧠 Go Concepts Demonstrated

| Concept | Where Used |
|---|---|
| Goroutines | Logger, state manager, interest ticker, tellers |
| Buffered channels | Transaction channel, rate limiter |
| Unbuffered channels | Log channel, response channel |
| Directional channels | Teller worker (`<-chan`, `chan<-`) |
| `select` statement | State manager, rate limiter, timeouts |
| `select` with `default` | Rate limiter refill |
| `time.NewTicker` | Interest accrual every 5 seconds |
| `time.After` | Transaction timeouts |
| Worker pool pattern | 3 tellers processing 10 random transactions |
| Bursty rate limiting | Pre-filled token bucket channel |
| `sync.Mutex` | Protecting transaction counter |
| `sync.WaitGroup` | Waiting for all tellers to finish |
| `defer` | Mutex unlocking, display balance cleanup |
| Stateful goroutine | Single goroutine owning the accounts map |
| `init()` | Pre-filling rate limiter tokens |

---

## 👥 Accounts

The simulation starts with 3 customers:

| ID | Name | Starting Balance |
|---|---|---|
| 1 | Alice | $1000.00 |
| 2 | Bob | $500.00 |
| 3 | Charlie | $750.00 |

---

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.21 or higher

### Run the game

```bash
# Clone or navigate to the project directory
cd bank-simulation

# Run directly
go run main.go

# Or build and run
go build -o bank main.go
./bank        # Linux/Mac
bank.exe      # Windows
```

---

## 🎮 How to Play

When you start the game you'll see this menu:

```
=============================================
     🏦  BANK SIMULATION CLI GAME
=============================================

  1. Deposit money
  2. Withdraw money
  3. Check balance
  4. Run automatic transactions (worker pool)
  5. View transaction count
  6. Exit
```

### Options

**1. Deposit money**
Enter a customer ID (1-3) and an amount to deposit into their account.

**2. Withdraw money**
Enter a customer ID and an amount to withdraw. If the customer has insufficient funds the transaction will be rejected.

**3. Check balance**
Enter a customer ID to view their current name and balance.

**4. Run automatic transactions**
Spawns a worker pool of 3 concurrent tellers that automatically process 10 random transactions. Watch the logger print everything happening in real time!

**5. View transaction count**
Displays the total number of transactions processed since the game started.

**6. Exit**
Shuts down all background goroutines and exits cleanly.

---

## 🔄 What Happens in the Background

While you're playing, several goroutines are running concurrently behind the scenes:

```
┌─────────────────────────────────────────────────┐
│                  main goroutine                  │
│              (handles user input)                │
└────────────────────┬────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
  loggerWorker  stateManager  interestTicker
  (prints logs) (owns accounts) (1% every 5s)
                     │
               refillRateLimiter
               (token every 500ms)
```

- **loggerWorker** — listens on the log channel and prints every transaction
- **stateManager** — the only goroutine that touches the accounts map, preventing race conditions
- **interestTicker** — fires every 5 seconds and adds 1% interest to all accounts
- **refillRateLimiter** — refills the rate limiter token bucket every 500ms

---

## 📁 Project Structure

```
bank-simulation/
└── main.go       # All code lives here
```

---

## 🗺️ Architecture Overview

```
User Input
    │
    ▼
main() ──── builds Transaction{} ──── transactionChannel ──── stateManager()
                                                                     │
                                                               updates accounts
                                                                     │
                                                            tx.Response channel
                                                                     │
                                                              back to main()
```

For the worker pool:
```
runWorkerPool()
    │
    ├── go teller(1, jobs, &wg)  ─┐
    ├── go teller(2, jobs, &wg)   ├── all consume from jobs channel concurrently
    └── go teller(3, jobs, &wg)  ─┘
              │
              ▼
    transactionChannel ──── stateManager()
```

---

## 📚 Learning Resources

This project was built following:
- [Go by Example](https://gobyexample.com) — the primary learning resource used throughout this project
- [The Go Tour](https://go.dev/tour) — official interactive Go tutorial
- [Effective Go](https://go.dev/doc/effective_go) — official Go best practices guide

---

## 🙏 Acknowledgements

Built as a hands-on learning project to understand Go concurrency patterns by porting a Python CLI simulation into idiomatic Go.
