package handlers

import (
    "fmt"
    "time"
    "log"
    "github.com/gin-gonic/gin"
    "bounty-system/internal/types"
    "bounty-system/internal/client"
)

type TaskHandler struct {
    blockchainClient *client.BlockchainClient
    tasks           map[string]types.Task
}

func NewTaskHandler() *TaskHandler {
    return &TaskHandler{
        blockchainClient: client.NewBlockchainClient(),
        tasks:           make(map[string]types.Task),
    }
}

// Admin middleware
func (h *TaskHandler) AdminRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        address := c.GetHeader("X-Wallet-Address")
        if address == "" {
            c.JSON(401, gin.H{"error": "wallet address required"})
            c.Abort()
            return
        }

        if !h.blockchainClient.adminManager.IsAdmin(address) {
            c.JSON(403, gin.H{"error": "admin access required"})
            c.Abort()
            return
        }

        c.Next()
    }
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
    tasks := make([]types.Task, 0, len(h.tasks))
    for _, task := range h.tasks {
        tasks = append(tasks, task)
    }
    c.JSON(200, tasks)
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
    var task types.Task
    if err := c.ShouldBindJSON(&task); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    log.Printf("Attempting to create task with creator: %s", task.Creator)
    
    // Validate address
    if !h.blockchainClient.ValidateAddress(task.Creator) {
        wallets := h.blockchainClient.GetTestWallets()
        c.JSON(400, gin.H{
            "error": "invalid creator address format",
            "message": "Address must start with 'serv1' and be 39-44 characters long",
            "valid_examples": wallets,
            "received_address": task.Creator,
            "received_length": len(task.Creator),
        })
        return
    }
    
    // Generate task ID
    task.ID = fmt.Sprintf("task-%d", time.Now().Unix())
    task.Status = "OPEN"
    
    // Store task
    h.tasks[task.ID] = task
    
    c.JSON(201, task)
}

func (h *TaskHandler) ClaimTask(c *gin.Context) {
    taskID := c.Param("id")
    
    task, exists := h.tasks[taskID]
    if !exists {
        c.JSON(404, gin.H{"error": "task not found"})
        return
    }
    
    if task.Status != "OPEN" {
        c.JSON(400, gin.H{"error": "task is not open for claiming"})
        return
    }
    
    var claim struct {
        Claimer string `json:"claimer"`
        Proof   string `json:"proof"`
    }
    
    if err := c.ShouldBindJSON(&claim); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Validate claimer address
    if !h.blockchainClient.ValidateAddress(claim.Claimer) {
        c.JSON(400, gin.H{
            "error": "invalid claimer address format",
            "message": "Address must start with 'serv1' and be 39-44 characters long",
        })
        return
    }
    
    task.Status = "CLAIMED"
    task.Claimer = claim.Claimer
    task.Proof = claim.Proof
    h.tasks[taskID] = task
    
    c.JSON(200, task)
}

func (h *TaskHandler) ApproveTask(c *gin.Context) {
    // Admin check is done by middleware
    taskID := c.Param("id")
    
    task, exists := h.tasks[taskID]
    if !exists {
        c.JSON(404, gin.H{"error": "task not found"})
        return
    }
    
    if task.Status != "CLAIMED" {
        c.JSON(400, gin.H{"error": "task must be claimed before approval"})
        return
    }
    
    task.Status = "COMPLETED"
    h.tasks[taskID] = task
    
    c.JSON(200, task)
}

func (h *TaskHandler) GetTasksByStatus(c *gin.Context) {
    status := c.Query("status")
    if status == "" {
        c.JSON(400, gin.H{"error": "status parameter is required"})
        return
    }
    
    var filteredTasks []types.Task
    for _, task := range h.tasks {
        if task.Status == status {
            filteredTasks = append(filteredTasks, task)
        }
    }
    
    c.JSON(200, filteredTasks)
}

// Admin only endpoints
func (h *TaskHandler) AddAdmin(c *gin.Context) {
    var req struct {
        Address string `json:"address"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    if !h.blockchainClient.ValidateAddress(req.Address) {
        c.JSON(400, gin.H{"error": "invalid address format"})
        return
    }
    
    h.blockchainClient.adminManager.AddAdmin(req.Address)
    c.JSON(200, gin.H{"message": "admin added successfully"})
}

func (h *TaskHandler) RemoveAdmin(c *gin.Context) {
    var req struct {
        Address string `json:"address"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    h.blockchainClient.adminManager.RemoveAdmin(req.Address)
    c.JSON(200, gin.H{"message": "admin removed successfully"})
}
