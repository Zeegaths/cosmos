package main

import (
    "log"
    "net/http"
    "encoding/json"
    "strings"
    "time"
    "fmt"
    "bounty-system/internal/client"
    intTypes "bounty-system/internal/types"
)

type Server struct {
    tasks map[string]intTypes.Task
    bc    *client.BlockchainClient
}

func NewServer() *Server {
    bc := client.NewBlockchainClient()
    
    // Get admin address
    adminAddr := bc.GetAdminAddress()
    
    log.Printf("\n=== IMPORTANT ADDRESSES ===")
    log.Printf("Admin Address: %s", adminAddr)
    log.Printf("Use this address in X-Wallet-Address header for admin operations")
    log.Printf("========================\n")

    return &Server{
        tasks: make(map[string]intTypes.Task),
        bc:    bc,
    }
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received request: %s %s", r.Method, r.URL.Path)
    
    switch {
    case r.Method == "GET" && r.URL.Path == "/tasks":
        s.handleGetTasks(w, r)
    case r.Method == "POST" && r.URL.Path == "/tasks":
        s.handleCreateTask(w, r)
    case r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "/claim"):
        s.handleClaimTask(w, r)
    case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/admin/tasks/"):
        s.handleApproveTask(w, r)
    case r.Method == "POST" && r.URL.Path == "/admin/admins":
        s.handleAdmins(w, r)
    case r.Method == "POST" && r.URL.Path == "/generate-address":
        s.handleGenerateAddress(w, r)
    case r.Method == "GET" && r.URL.Path == "/addresses":
        s.handleListAddresses(w, r)
    default:
        log.Printf("No route match found for: %s %s", r.Method, r.URL.Path)
        http.NotFound(w, r)
    }
}

func (s *Server) handleListAddresses(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    addresses := s.bc.ListAddresses()
    json.NewEncoder(w).Encode(map[string]interface{}{
        "admin_address": s.bc.GetAdminAddress(),
        "all_addresses": addresses,
    })
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var task intTypes.Task
    if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    task.ID = fmt.Sprintf("task-%d", time.Now().Unix())
    task.Status = "OPEN"

    if err := s.bc.CreateTask(task); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    s.tasks[task.ID] = task
    log.Printf("Created task: %s", task.ID)
    json.NewEncoder(w).Encode(task)
}

func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    tasks, err := s.bc.ListTasks()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(tasks)
}

func (s *Server) handleGenerateAddress(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    var req struct {
        Seed string `json:"seed"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    address := s.bc.GenerateTestAddress(req.Seed)
    if address == "" {
        http.Error(w, "Failed to generate address", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{
        "address": address,
    })
}

func (s *Server) handleClaimTask(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 4 {
        http.Error(w, "Invalid URL format", http.StatusBadRequest)
        return
    }
    taskID := parts[2]

    var claim struct {
        Claimer string `json:"claimer"`
        Proof   string `json:"proof"`
    }
    if err := json.NewDecoder(r.Body).Decode(&claim); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := s.bc.ClaimTask(taskID, claim.Claimer, claim.Proof); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Get updated task
    tasks, _ := s.bc.ListTasks()
    var claimedTask intTypes.Task
    for _, t := range tasks {
        if t.ID == taskID {
            claimedTask = t
            break
        }
    }

    json.NewEncoder(w).Encode(claimedTask)
}

func (s *Server) handleApproveTask(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    adminAddr := r.Header.Get("X-Wallet-Address")
    if !s.bc.IsAdmin(adminAddr) {
        http.Error(w, "Unauthorized - Admin access required", http.StatusUnauthorized)
        return
    }

    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 4 {
        http.Error(w, "Invalid URL format", http.StatusBadRequest)
        return
    }
    taskID := parts[3]

    tasks, _ := s.bc.ListTasks()
    var task intTypes.Task
    for _, t := range tasks {
        if t.ID == taskID {
            task = t
            break
        }
    }

    if task.ID == "" {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }

    if err := s.bc.ApproveTask(task, adminAddr); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Get updated task
    tasks, _ = s.bc.ListTasks()
    for _, t := range tasks {
        if t.ID == taskID {
            task = t
            break
        }
    }

    json.NewEncoder(w).Encode(task)
}

func (s *Server) handleAdmins(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    adminAddr := r.Header.Get("X-Wallet-Address")
    if !s.bc.IsAdmin(adminAddr) {
        http.Error(w, "Unauthorized - Admin access required", http.StatusUnauthorized)
        return
    }

    var req struct {
        Address string `json:"address"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := s.bc.AddAdmin(req.Address, adminAddr); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]string{"message": "Admin added successfully"})
}

func main() {
    server := NewServer()
    
    log.Printf("Starting Tokenized Task Bounty System...")
    log.Printf("Chain ID: %s", server.bc.GetChainID())
    log.Printf("RPC Endpoint: %s", server.bc.GetRPCEndpoint())
    log.Printf("REST Endpoint: %s", server.bc.GetRESTEndpoint())
    
    log.Printf("\nServer starting on :8080")
    log.Printf("Available endpoints:")
    log.Printf("GET  /addresses        - List all addresses")
    log.Printf("POST /generate-address - Generate a new address")
    log.Printf("POST /tasks           - Create a task")
    log.Printf("GET  /tasks           - List all tasks")
    log.Printf("PUT  /tasks/{id}/claim- Claim a task")
    log.Printf("PUT  /admin/tasks/{id}- Approve a task")
    
    log.Fatal(http.ListenAndServe(":8080", server))
}