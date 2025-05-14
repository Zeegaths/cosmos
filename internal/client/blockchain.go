package client

import (
   "fmt"
   "encoding/json"
   "net/http"
   "bytes"
   "log"
   "time"
   "bounty-system/internal/types"
)

type BlockchainClient struct {
   rpcEndpoint  string
   restEndpoint string
   chainID      string
}

func NewBlockchainClient() *BlockchainClient {
   return &BlockchainClient{
       rpcEndpoint:  "https://node0.testnet.knowfreedom.io:26657",
       restEndpoint: "https://node0.testnet.knowfreedom.io:1317",
       chainID:      "kfchain",
   }
}

func (c *BlockchainClient) CreateTask(task types.Task) error {
   tx := struct {
       Body struct {
           Messages []struct {
               Type   string `json:"@type"`
               From   string `json:"from_address"`
               To     string `json:"to_address"`
               Amount []struct {
                   Denom  string `json:"denom"`
                   Amount string `json:"amount"`
               } `json:"amount"`
           } `json:"messages"`
           Memo string `json:"memo"`
       } `json:"body"`
   }{}

   // Set up transaction
   tx.Body.Messages = []struct {
       Type   string `json:"@type"`
       From   string `json:"from_address"`
       To     string `json:"to_address"`
       Amount []struct {
           Denom  string `json:"denom"`
           Amount string `json:"amount"`
       } `json:"amount"`
   }{
       {
           Type: "/cosmos.bank.v1beta1.MsgSend",
           From: task.Creator,
           To:   "serv1escrow",
           Amount: []struct {
               Denom  string `json:"denom"`
               Amount string `json:"amount"`
           }{
               {
                   Denom:  "microSERVDR",
                   Amount: task.Bounty,
               },
           },
       },
   }

   // Store task data in memo
   taskData, err := json.Marshal(task)
   if err != nil {
       return fmt.Errorf("failed to marshal task: %v", err)
   }
   tx.Body.Memo = string(taskData)

   return c.submitTransaction(tx)
}

func (c *BlockchainClient) GetTask(taskID string) (*types.Task, error) {
   url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?events=tx.memo='%s'", 
       c.restEndpoint, taskID)
   
   resp, err := http.Get(url)
   if err != nil {
       return nil, fmt.Errorf("failed to get task: %v", err)
   }
   defer resp.Body.Close()

   var result struct {
       Txs []struct {
           Body struct {
               Memo string `json:"memo"`
           } `json:"body"`
       } `json:"txs"`
   }

   if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
       return nil, fmt.Errorf("failed to decode response: %v", err)
   }

   if len(result.Txs) == 0 {
       return nil, nil
   }

   var task types.Task
   if err := json.Unmarshal([]byte(result.Txs[0].Body.Memo), &task); err != nil {
       return nil, fmt.Errorf("failed to unmarshal task: %v", err)
   }

   return &task, nil
}

func (c *BlockchainClient) ListTasks() ([]types.Task, error) {
   url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?events=tx.memo CONTAINS 'task-'",
       c.restEndpoint)
   
   resp, err := http.Get(url)
   if err != nil {
       return nil, fmt.Errorf("failed to list tasks: %v", err)
   }
   defer resp.Body.Close()

   var result struct {
       Txs []struct {
           Body struct {
               Memo string `json:"memo"`
           } `json:"body"`
       } `json:"txs"`
   }

   if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
       return nil, fmt.Errorf("failed to decode response: %v", err)
   }

   tasks := make([]types.Task, 0)
   for _, tx := range result.Txs {
       var task types.Task
       if err := json.Unmarshal([]byte(tx.Body.Memo), &task); err != nil {
           continue // Skip invalid task data
       }
       tasks = append(tasks, task)
   }

   return tasks, nil
}

func (c *BlockchainClient) ReleaseBounty(task types.Task) error {
   tx := struct {
       Body struct {
           Messages []struct {
               Type   string `json:"@type"`
               From   string `json:"from_address"`
               To     string `json:"to_address"`
               Amount []struct {
                   Denom  string `json:"denom"`
                   Amount string `json:"amount"`
               } `json:"amount"`
           } `json:"messages"`
           Memo string `json:"memo"`
       } `json:"body"`
   }{}

   tx.Body.Messages = []struct {
       Type   string `json:"@type"`
       From   string `json:"from_address"`
       To     string `json:"to_address"`
       Amount []struct {
           Denom  string `json:"denom"`
           Amount string `json:"amount"`
       } `json:"amount"`
   }{
       {
           Type: "/cosmos.bank.v1beta1.MsgSend",
           From: "serv1escrow",
           To:   task.Claimer,
           Amount: []struct {
               Denom  string `json:"denom"`
               Amount string `json:"amount"`
           }{
               {
                   Denom:  "microSERVDR",
                   Amount: task.Bounty,
               },
           },
       },
   }

   task.Status = "COMPLETED"
   taskData, err := json.Marshal(task)
   if err != nil {
       return fmt.Errorf("failed to marshal task: %v", err)
   }
   tx.Body.Memo = string(taskData)

   return c.submitTransaction(tx)
}

func (c *BlockchainClient) submitTransaction(tx interface{}) error {
   txBytes, err := json.Marshal(tx)
   if err != nil {
       return fmt.Errorf("failed to marshal transaction: %v", err)
   }

   resp, err := http.Post(
       fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", c.restEndpoint),
       "application/json",
       bytes.NewBuffer(txBytes),
   )
   if err != nil {
       return fmt.Errorf("failed to submit transaction: %v", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != 200 {
       body, _ := ioutil.ReadAll(resp.Body)
       return fmt.Errorf("transaction failed: %s: %s", resp.Status, string(body))
   }

   return nil
}

func (c *BlockchainClient) GetBalance(address string) (string, error) {
   url := fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", c.restEndpoint, address)
   
   resp, err := http.Get(url)
   if err != nil {
       return "", fmt.Errorf("failed to get balance: %v", err)
   }
   defer resp.Body.Close()

   var result struct {
       Balances []struct {
           Denom  string `json:"denom"`
           Amount string `json:"amount"`
       } `json:"balances"`
   }

   if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
       return "", fmt.Errorf("failed to decode response: %v", err)
   }

   for _, balance := range result.Balances {
       if balance.Denom == "microSERVDR" {
           return balance.Amount, nil
       }
   }

   return "0", nil
}

func (c *BlockchainClient) GetTestWallets() []string {
   return []string{
       "serv1ph3trl8q0mtc6dfz5xsqq6pksl7w7rxx0hk3q",
       "serv1mj9k8h7l6n5m4p3q2r1s0t9v8w7x6y5z4a3b",
   }
}
