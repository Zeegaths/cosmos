package main

import (
    "log"
    "bounty-system/internal/client"
    "bounty-system/internal/types"
)

func main() {
    c := client.NewBlockchainClient()
    
    // Get test wallets
    wallets := c.GetTestWallets()
    log.Printf("Test wallets: %v", wallets)
    
    // Check balance
    balance, err := c.GetBalance(wallets[0])
    if err != nil {
        log.Fatalf("Error checking balance: %v", err)
    }
    log.Printf("Balance for %s: %s", wallets[0], balance)
    
    // Try to create a task
    task := types.Task{
        Title:       "Test Task",
        Description: "Test Description",
        Creator:     wallets[0],
        Bounty:      "1000000microSERVDR",
        Status:      "OPEN",
    }
    
    err = c.CreateTask(task)
    if err != nil {
        log.Printf("Error creating task: %v", err)
    } else {
        log.Printf("Task created successfully")
    }
}
