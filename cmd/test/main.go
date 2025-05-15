package main

import (
    "log"
    "fmt"
    "time"
    "bounty-system/internal/client"
    "bounty-system/internal/types"
)

func main() {
    c := client.NewBlockchainClient()
    
    // Get test wallets
    wallets := c.GetTestWallets()
    log.Printf("Test wallets: %v", wallets)
    
    // Check admin status
    for _, wallet := range wallets {
        isAdmin := c.IsAdmin(wallet)
        log.Printf("Wallet %s is admin: %v", wallet, isAdmin)
    }
    
    // Create task with test user wallet
    taskID := fmt.Sprintf("task-%d", time.Now().Unix())
    task := types.Task{
        ID:          taskID,
        Title:       "Test Task",
        Description: "Test Description",
        Creator:     wallets[1], // Use non-admin wallet
        Bounty:      "1000000", // Amount without denom
        Status:      "OPEN",
    }
    
    err := c.CreateTask(task)
    if err != nil {
        log.Printf("Error creating task: %v", err)
    } else {
        log.Printf("Task created successfully")
        
        // Try to approve with non-admin
        err = c.ApproveTask(task, wallets[1])
        if err != nil {
            log.Printf("Expected error with non-admin approval: %v", err)
        }
        
        // Try to approve with admin
        task.Status = "CLAIMED"
        task.Claimer = wallets[1]
        err = c.ApproveTask(task, wallets[0])
        if err != nil {
            log.Printf("Error approving task with admin: %v", err)
        } else {
            log.Printf("Task approved successfully by admin")
        }
    }
    
    // Test admin management
    newAdmin := "serv1newadmin..."
    
    // Try adding admin with non-admin wallet
    err = c.AddAdmin(newAdmin, wallets[1])
    if err != nil {
        log.Printf("Expected error when non-admin adds admin: %v", err)
    }
    
    // Add admin with admin wallet
    err = c.AddAdmin(newAdmin, wallets[0])
    if err != nil {
        log.Printf("Error adding new admin: %v", err)
    } else {
        log.Printf("New admin added successfully")
    }
    
    // List all admins
    admins := c.ListAdmins()
    log.Printf("Current admins: %v", admins)
}
