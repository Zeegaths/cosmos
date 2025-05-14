package client

import (
    "sync"
    "bounty-system/internal/types"
)

type AdminManager struct {
    admins map[string]bool
    mu     sync.RWMutex
}

func NewAdminManager() *AdminManager {
    return &AdminManager{
        admins: make(map[string]bool),
    }
}

func (am *AdminManager) AddAdmin(address string) {
    am.mu.Lock()
    defer am.mu.Unlock()
    am.admins[address] = true
}

func (am *AdminManager) RemoveAdmin(address string) {
    am.mu.Lock()
    defer am.mu.Unlock()
    delete(am.admins, address)
}

func (am *AdminManager) IsAdmin(address string) bool {
    am.mu.RLock()
    defer am.mu.RUnlock()
    return am.admins[address]
}
