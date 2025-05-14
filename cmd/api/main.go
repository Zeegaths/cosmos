package main

import (
   "net/http"
   "encoding/json"
   "log"
   "strings"
   "time"
   "fmt"
)

type Task struct {
   ID          string `json:"id"`
   Title       string `json:"title"`
   Description string `json:"description"`
   Creator     string `json:"creator"`
   Bounty      string `json:"bounty"`
   Status      string `json:"status"`
   Claimer     string `json:"claimer,omitempty"`
   Proof       string `json:"proof,omitempty"`
}

type Server struct {
   tasks          map[string]Task
   adminAddresses map[string]bool
}

func NewServer() *Server {
   s := &Server{
       tasks:          make(map[string]Task),
       adminAddresses: make(map[string]bool),
   }
   // Add initial admin
   s.adminAddresses["serv1ph3trl8q0mtc6dfz5xsqq6pksl7w7rxx0hk3q"] = true
   return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   log.Printf("Received request: %s %s", r.Method, r.URL.Path)
   
   switch {
   case r.Method == "GET" && r.URL.Path == "/tasks":
       s.handleListTasks(w, r)
   case r.Method == "POST" && r.URL.Path == "/tasks":
       s.handleCreateTask(w, r)
   case r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "/claim"):
       s.handleClaimTask(w, r)
   case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/admin/tasks/"):
       s.handleApproveTask(w, r)
   case r.Method == "POST" && r.URL.Path == "/admin/admins":
       s.handleAdmins(w, r)
   default:
       log.Printf("No route match found for: %s %s", r.Method, r.URL.Path)
       http.NotFound(w, r)
   }
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
   w.Header().Set("Content-Type", "application/json")
   tasks := make([]Task, 0, len(s.tasks))
   for _, task := range s.tasks {
       tasks = append(tasks, task)
   }
   json.NewEncoder(w).Encode(tasks)
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
   w.Header().Set("Content-Type", "application/json")
   var task Task
   if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
       http.Error(w, err.Error(), http.StatusBadRequest)
       return
   }
   
   task.ID = fmt.Sprintf("task-%d", time.Now().Unix())
   task.Status = "OPEN"
   s.tasks[task.ID] = task
   
   log.Printf("Created task: %s", task.ID)
   json.NewEncoder(w).Encode(task)
}

func (s *Server) handleClaimTask(w http.ResponseWriter, r *http.Request) {
   w.Header().Set("Content-Type", "application/json")
   
   // Extract task ID from URL: /tasks/{id}/claim
   parts := strings.Split(r.URL.Path, "/")
   if len(parts) < 4 {
       http.Error(w, "Invalid URL format", http.StatusBadRequest)
       return
   }
   taskID := parts[2]
   
   task, exists := s.tasks[taskID]
   if !exists {
       http.Error(w, "Task not found", http.StatusNotFound)
       return
   }

   if task.Status != "OPEN" {
       http.Error(w, "Task is not open for claiming", http.StatusBadRequest)
       return
   }

   var claim struct {
       Claimer string `json:"claimer"`
       Proof   string `json:"proof"`
   }
   if err := json.NewDecoder(r.Body).Decode(&claim); err != nil {
       http.Error(w, err.Error(), http.StatusBadRequest)
       return
   }

   task.Status = "CLAIMED"
   task.Claimer = claim.Claimer
   task.Proof = claim.Proof
   s.tasks[taskID] = task

   log.Printf("Task %s claimed by %s", taskID, claim.Claimer)
   json.NewEncoder(w).Encode(task)
}

func (s *Server) handleApproveTask(w http.ResponseWriter, r *http.Request) {
   w.Header().Set("Content-Type", "application/json")

   // Check admin authorization
   adminAddr := r.Header.Get("X-Wallet-Address")
   if !s.adminAddresses[adminAddr] {
       http.Error(w, "Unauthorized - Admin access required", http.StatusUnauthorized)
       return
   }

   // Extract task ID from URL: /admin/tasks/{id}
   parts := strings.Split(r.URL.Path, "/")
   if len(parts) < 4 {
       http.Error(w, "Invalid URL format", http.StatusBadRequest)
       return
   }
   taskID := parts[3]

   task, exists := s.tasks[taskID]
   if !exists {
       http.Error(w, "Task not found", http.StatusNotFound)
       return
   }

   if task.Status != "CLAIMED" {
       http.Error(w, "Task must be claimed before approval", http.StatusBadRequest)
       return
   }

   task.Status = "COMPLETED"
   s.tasks[taskID] = task

   log.Printf("Task %s approved by admin %s", taskID, adminAddr)
   json.NewEncoder(w).Encode(task)
}

func (s *Server) handleAdmins(w http.ResponseWriter, r *http.Request) {
   w.Header().Set("Content-Type", "application/json")

   adminAddr := r.Header.Get("X-Wallet-Address")
   if !s.adminAddresses[adminAddr] {
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

   s.adminAddresses[req.Address] = true
   
   log.Printf("New admin added: %s", req.Address)
   json.NewEncoder(w).Encode(map[string]string{"message": "Admin added successfully"})
}

func main() {
   server := NewServer()
   
   log.Printf("Server starting on :8080")
   log.Fatal(http.ListenAndServe(":8080", server))
}
