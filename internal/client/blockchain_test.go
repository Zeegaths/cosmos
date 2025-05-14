package client

import (
    "testing"
    "log"
)

func TestBlockchainClient(t *testing.T) {
    client := NewBlockchainClient()
    
    // Test wallet validation
    testWallets := client.GetTestWallets()
    for _, wallet := range testWallets {
        balance, err := client.GetBalance(wallet)
        if err != nil {
            log.Printf("Error getting balance for %s: %v", wallet, err)
        } else {
            log.Printf("Wallet %s balance: %s", wallet, balance)
        }
    }
}
