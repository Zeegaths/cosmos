package client

import (
    "fmt"
    "log"       
    intTypes "bounty-system/internal/types"
    "crypto/sha256"
    "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type BlockchainClient struct {
    tasks          map[string]intTypes.Task
    adminWallets   map[string]bool
    walletKeys     map[string]string
    adminAddress   string            // Store the admin address
}

func NewBlockchainClient() *BlockchainClient {
    client := &BlockchainClient{
        tasks:          make(map[string]intTypes.Task),
        adminWallets:   make(map[string]bool),
        walletKeys:     make(map[string]string),
    }
    
    // Generate initial admin address
    adminAddr := client.GenerateTestAddress("admin-1")
    client.adminWallets[adminAddr] = true
    client.adminAddress = adminAddr
    
    log.Printf("Created admin wallet: %s", adminAddr)
    
    return client
}

func (c *BlockchainClient) GetAdminAddress() string {
    return c.adminAddress
}

func (c *BlockchainClient) GenerateTestAddress(seed string) string {
    hasher := sha256.New()
    hasher.Write([]byte(seed))
    hash := hasher.Sum(nil)
    
    privKey := secp256k1.PrivKey{Key: hash}
    addr := sdk.AccAddress(privKey.PubKey().Address())
    bech32Addr := addr.String()
    
    // Store the address
    c.walletKeys[bech32Addr] = seed
    log.Printf("Generated address from seed '%s': %s", seed, bech32Addr)
    
    return bech32Addr
}

func (c *BlockchainClient) CreateTask(task intTypes.Task) error {
    if task.ID == "" || task.Title == "" || task.Bounty == "" {
        return fmt.Errorf("invalid task parameters")
    }
    
    // Store task in memory
    c.tasks[task.ID] = task
    log.Printf("Created task: %+v", task)
    return nil
}

func (c *BlockchainClient) ListTasks() ([]intTypes.Task, error) {
    tasks := make([]intTypes.Task, 0, len(c.tasks))
    for _, task := range c.tasks {
        tasks = append(tasks, task)
    }
    return tasks, nil
}

func (c *BlockchainClient) ClaimTask(taskID string, claimer string, proof string) error {
    task, exists := c.tasks[taskID]
    if !exists {
        return fmt.Errorf("task not found")
    }
    
    if task.Status != "OPEN" {
        return fmt.Errorf("task is not open for claiming")
    }
    
    task.Status = "CLAIMED"
    task.Claimer = claimer
    task.Proof = proof
    c.tasks[taskID] = task
    
    log.Printf("Task %s claimed by %s", taskID, claimer)
    return nil
}

func (c *BlockchainClient) ApproveTask(task intTypes.Task, approver string) error {
    if !c.IsAdmin(approver) {
        return fmt.Errorf("only admins can approve tasks")
    }
    
    existingTask, exists := c.tasks[task.ID]
    if !exists {
        return fmt.Errorf("task not found")
    }
    
    if existingTask.Status != "CLAIMED" {
        return fmt.Errorf("task must be claimed before approval")
    }
    
    existingTask.Status = "COMPLETED"
    c.tasks[task.ID] = existingTask
    
    log.Printf("Task %s approved by admin %s", task.ID, approver)
    return nil
}

func (c *BlockchainClient) IsAdmin(address string) bool {
    return c.adminWallets[address]
}

func (c *BlockchainClient) GetChainID() string {
    return "mock-chain"
}

func (c *BlockchainClient) GetRPCEndpoint() string {
    return "mock://localhost:26657"
}

func (c *BlockchainClient) GetRESTEndpoint() string {
    return "mock://localhost:1317"
}

func (c *BlockchainClient) ValidateAddress(address string) bool {
    // In mock mode, just check if we've generated this address
    _, exists := c.walletKeys[address]
    return exists
}

func (c *BlockchainClient) ListAddresses() []string {
    addresses := make([]string, 0, len(c.walletKeys))
    for addr := range c.walletKeys {
        addresses = append(addresses, addr)
    }
    return addresses
}

// Admin management functions
func (c *BlockchainClient) AddAdmin(address string, requestor string) error {
    if !c.IsAdmin(requestor) {
        return fmt.Errorf("only admins can add new admins")
    }
    c.adminWallets[address] = true
    return nil
}

func (c *BlockchainClient) RemoveAdmin(address string, requestor string) error {
    if !c.IsAdmin(requestor) {
        return fmt.Errorf("only admins can remove admins")
    }
    if len(c.adminWallets) <= 1 {
        return fmt.Errorf("cannot remove last admin")
    }
    delete(c.adminWallets, address)
    return nil
}

func (c *BlockchainClient) ListAdmins() []string {
    admins := make([]string, 0, len(c.adminWallets))
    for admin := range c.adminWallets {
        admins = append(admins, admin)
    }
    return admins
}








// package client

// import (
//     "fmt"
//     "encoding/json"
//     "net/http"
//     "bytes"
//     "io/ioutil"
//     "log"
//     "time"    
//     intTypes "bounty-system/internal/types"
//     "encoding/base64"
//     "crypto/sha256"
//     "strings"
    
//     "github.com/cosmos/cosmos-sdk/codec"
//     sdkTypes "github.com/cosmos/cosmos-sdk/codec/types"  // Aliased to sdkTypes
//     sdk "github.com/cosmos/cosmos-sdk/types"
//     banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
//     "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
//     cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
// )

// type Transaction struct {
//     Body       TxBody     `json:"body"`
//     AuthInfo   AuthInfo   `json:"auth_info"`
//     Signatures []string   `json:"signatures"`
// }

// type TxBody struct {
//     Messages      []json.RawMessage `json:"messages"`
//     Memo          string           `json:"memo"`
//     TimeoutHeight uint64           `json:"timeout_height"`
// }

// type AuthInfo struct {
//     SignerInfos []SignerInfo `json:"signer_infos"`
//     Fee         Fee          `json:"fee"`
// }

// type SignerInfo struct {
//     PublicKey PublicKey `json:"public_key"`
//     ModeInfo  ModeInfo  `json:"mode_info"`
//     Sequence  string    `json:"sequence"`
// }

// type PublicKey struct {
//     Type  string `json:"@type"`
//     Key   string `json:"key"`
// }

// type ModeInfo struct {
//     Single Single `json:"single"`
// }

// type Single struct {
//     Mode string `json:"mode"`
// }

// type Fee struct {
//     Amount   []Coin `json:"amount"`
//     GasLimit uint64 `json:"gas_limit"`
// }

// type Coin struct {
//     Denom  string `json:"denom"`
//     Amount string `json:"amount"`
// }

// func init() {
//     config := sdk.GetConfig()
    
//     // Set prefixes - the key here is that we need to use compatible charset
//     // Cosmos SDK uses Bech32 encoding with specific charset rules
//     // "serv" prefix should work with proper Bech32 encoding
//     config.SetBech32PrefixForAccount("serv", "servpub")
//     config.SetBech32PrefixForValidator("servvaloper", "servvaloperpub")
//     config.SetBech32PrefixForConsensusNode("servvalcons", "servvalconspub")
    
//     // Seal the config to prevent further modifications
//     config.Seal()
// }

// type EncodingConfig struct {
//     InterfaceRegistry sdkTypes.InterfaceRegistry  // Changed to sdkTypes
//     Marshaler         codec.Codec
//     Amino            *codec.LegacyAmino
// }

// func MakeEncodingConfig() EncodingConfig {
//     amino := codec.NewLegacyAmino()
//     interfaceRegistry := sdkTypes.NewInterfaceRegistry()  // Changed to sdkTypes
//     marshaler := codec.NewProtoCodec(interfaceRegistry)
    
//     // Register crypto interfaces
//     cryptocodec.RegisterInterfaces(interfaceRegistry)
    
//     // Register your custom types
//     banktypes.RegisterInterfaces(interfaceRegistry)
    
//     return EncodingConfig{
//         InterfaceRegistry: interfaceRegistry,
//         Marshaler:        marshaler,
//         Amino:           amino,
//     }
// }

// // Constants for task status
// const (
//     TaskStatusOpen     = "OPEN"
//     TaskStatusClaimed  = "CLAIMED"
//     TaskStatusApproved = "APPROVED"
//     TaskStatusRejected = "REJECTED"
// )

// // Account info structures
// type AccountInfo struct {
//     Account struct {
//         Address       string `json:"address"`
//         AccountNumber string `json:"account_number"`
//         Sequence     string `json:"sequence"`
//         PubKey       struct {
//             Type  string `json:"@type"`
//             Key   string `json:"key"`
//         } `json:"pub_key"`
//     } `json:"account"`
// }


// // Wallet structures
// type WalletKey struct {
//     Address    string
//     PubKey     string
//     PrivKey    string
// }

// // BlockchainClient structure
// type BlockchainClient struct {
//     rpcEndpoint    string
//     restEndpoint   string
//     chainID        string
//     adminWallets   map[string]bool
//     walletKeys     map[string]WalletKey
//     encodingConfig EncodingConfig
// }

// func NewBlockchainClient() *BlockchainClient {
//     // Create encoding config
//     encConfig := MakeEncodingConfig()

//     client := &BlockchainClient{
//         rpcEndpoint:    "https://node0.testnet.knowfreedom.io:26657",
//         restEndpoint:   "https://node0.testnet.knowfreedom.io:1317",
//         chainID:        "kfchain",
//         adminWallets:   make(map[string]bool),
//         walletKeys:     make(map[string]WalletKey),
//         encodingConfig: encConfig,
//     }
    
//     // Generate a deterministic admin address
//     adminWallet := client.GenerateTestAddress("admin-1")
//     client.adminWallets[adminWallet] = true
    
//     // Create a deterministic private key for the admin wallet
//     adminPrivKeyBytes := make([]byte, 32)
//     adminHasher := sha256.New()
//     adminHasher.Write([]byte("admin-key-1"))
//     copy(adminPrivKeyBytes, adminHasher.Sum(nil))
    
//     // Generate the public key
//     privKey := secp256k1.PrivKey{Key: adminPrivKeyBytes}
//     pubKey := privKey.PubKey()
//     pubKeyBytes := pubKey.Bytes()
    
//     // Store the wallet info
//     client.walletKeys[adminWallet] = WalletKey{
//         Address: adminWallet,
//         PubKey:  base64.StdEncoding.EncodeToString(pubKeyBytes),
//         PrivKey: base64.StdEncoding.EncodeToString(adminPrivKeyBytes),
//     }
    
//     // Generate a secondary non-admin wallet
//     secondaryWallet := client.GenerateTestAddress("user-1")
    
//     // Create a deterministic private key for the secondary wallet
//     secondaryPrivKeyBytes := make([]byte, 32)
//     secondaryHasher := sha256.New()
//     secondaryHasher.Write([]byte("user-key-1"))
//     copy(secondaryPrivKeyBytes, secondaryHasher.Sum(nil))
    
//     // Generate the public key for secondary wallet
//     secondaryPrivKey := secp256k1.PrivKey{Key: secondaryPrivKeyBytes}
//     secondaryPubKey := secondaryPrivKey.PubKey()
//     secondaryPubKeyBytes := secondaryPubKey.Bytes()
    
//     client.walletKeys[secondaryWallet] = WalletKey{
//         Address: secondaryWallet,
//         PubKey:  base64.StdEncoding.EncodeToString(secondaryPubKeyBytes),
//         PrivKey: base64.StdEncoding.EncodeToString(secondaryPrivKeyBytes),
//     }

//     return client
// }

// func (c *BlockchainClient) GenerateTestAddress(seed string) string {
//     // Use crypto/sha256 package for deterministic address generation
//     hasher := sha256.New()
//     hasher.Write([]byte(seed))
//     hash := hasher.Sum(nil)
    
//     // Create an AccAddress from the first 20 bytes of the hash (standard for Cosmos addresses)
//     addr := sdk.AccAddress(hash[:20])
    
//     // This will automatically use the "serv" prefix configured in init()
//     bech32Addr := addr.String()
    
//     // Verify the generated address
//     if !strings.HasPrefix(bech32Addr, "serv1") {
//         log.Printf("Warning: Generated address doesn't have correct prefix: %s", bech32Addr)
//         return ""
//     }
    
//     log.Printf("Generated valid address from seed '%s': %s", seed, bech32Addr)
//     return bech32Addr
// }

// func (c *BlockchainClient) CreateTask(task intTypes.Task) error {
//     // Validate the task
//     if task.ID == "" {
//         return fmt.Errorf("task ID cannot be empty")
//     }
    
//     if task.Title == "" {
//         return fmt.Errorf("task title cannot be empty")
//     }
    
//     if task.Bounty == "" {
//         return fmt.Errorf("task bounty cannot be empty")
//     }
    
//     // Validate creator address
//     if !c.ValidateAddress(task.Creator) {
//         return fmt.Errorf("invalid creator address: %s", task.Creator)
//     }
    
//     // Convert addresses to bech32
//     fromAddr, err := sdk.AccAddressFromBech32(task.Creator)
//     if err != nil {
//         return fmt.Errorf("invalid creator address: %v", err)
//     }

//     // For escrow, we can use a derived address or a pre-defined one
//     // Here we're using a constant for simplicity
//     escrowAddr := "serv1escrow000000000000000000000000000000"
//     toAddr, err := sdk.AccAddressFromBech32(escrowAddr)
//     if err != nil {
//         // If the constant doesn't work, derive a deterministic escrow address
//         hasher := sha256.New()
//         hasher.Write([]byte("escrow"))
//         hash := hasher.Sum(nil)
//         toAddr = sdk.AccAddress(hash[:20])
//         escrowAddr = toAddr.String()
//     }

//     // Create the message
//     msg := map[string]interface{}{
//         "@type":        "/cosmos.bank.v1beta1.MsgSend",
//         "from_address": fromAddr.String(),
//         "to_address":   escrowAddr,
//         "amount": []map[string]string{
//             {
//                 "denom":  "microSERVDR",
//                 "amount": task.Bounty,
//             },
//         },
//     }

//     // Add task metadata
//     metadata := map[string]interface{}{
//         "type":        "create_task",
//         "title":       task.Title,
//         "description": task.Description,
//         "status":      "OPEN",
//         "task_id":     task.ID,
//     }

//     metadataBytes, err := json.Marshal(metadata)
//     if err != nil {
//         return fmt.Errorf("failed to marshal metadata: %v", err)
//     }

//     msgBytes, err := json.Marshal(msg)
//     if err != nil {
//         return fmt.Errorf("failed to marshal message: %v", err)
//     }

//     // Get account sequence
//     accountInfo, err := c.getAccountInfo(task.Creator)
//     if err != nil {
//         log.Printf("Warning: Failed to get account info, using default sequence: %v", err)
//         // Continue with default sequence
//     }

//     // Create transaction
//     tx := Transaction{
//         Body: TxBody{
//             Messages:      []json.RawMessage{msgBytes},
//             Memo:          string(metadataBytes),
//             TimeoutHeight: uint64(time.Now().Unix() + 300), // 5 minutes from now
//         },
//         AuthInfo: AuthInfo{
//             SignerInfos: []SignerInfo{
//                 {
//                     PublicKey: PublicKey{
//                         Type: "/cosmos.crypto.secp256k1.PubKey",
//                         Key:  c.walletKeys[task.Creator].PubKey,
//                     },
//                     ModeInfo: ModeInfo{
//                         Single: Single{
//                             Mode: "SIGN_MODE_DIRECT",
//                         },
//                     },
//                     Sequence: accountInfo.Account.Sequence,
//                 },
//             },
//             Fee: Fee{
//                 Amount: []Coin{
//                     {
//                         Denom:  "microSERVDR",
//                         Amount: "1000", // Fixed gas fee
//                     },
//                 },
//                 GasLimit: 200000, // Standard gas limit
//             },
//         },
//         Signatures: make([]string, 0),
//     }

//     // Sign the transaction
//     if err := c.signTransaction(&tx, task.Creator); err != nil {
//         return fmt.Errorf("failed to sign transaction: %v", err)
//     }

//     // Marshal and submit the transaction
//     txBytes, err := json.Marshal(tx)
//     if err != nil {
//         return fmt.Errorf("failed to marshal transaction: %v", err)
//     }
    
//     return c.submitTransaction(txBytes)
// }

// // Helper method to get account info including sequence number
// func (c *BlockchainClient) getAccountInfo(address string) (AccountInfo, error) {
//     var accountInfo AccountInfo
    
//     url := fmt.Sprintf("%s/cosmos/auth/v1beta1/accounts/%s", c.restEndpoint, address)
//     resp, err := http.Get(url)
//     if err != nil {
//         return accountInfo, fmt.Errorf("failed to get account info: %v", err)
//     }
//     defer resp.Body.Close()
    
//     body, err := ioutil.ReadAll(resp.Body)
//     if err != nil {
//         return accountInfo, fmt.Errorf("failed to read response: %v", err)
//     }
    
//     if resp.StatusCode != http.StatusOK {
//         return accountInfo, fmt.Errorf("failed to get account info, status: %d, response: %s", 
//             resp.StatusCode, string(body))
//     }
    
//     if err := json.Unmarshal(body, &accountInfo); err != nil {
//         return accountInfo, fmt.Errorf("failed to parse account info: %v", err)
//     }
    
//     return accountInfo, nil
// }
// // Update ApproveTask to use proper message format
// func (c *BlockchainClient) ApproveTask(task intTypes.Task, approver string) error {
//     if !c.IsAdmin(approver) {
//         return fmt.Errorf("only admins can approve tasks")
//     }

//     msg := map[string]interface{}{
//         "@type": "/cosmos.bank.v1beta1.MsgSend",
//         "from_address": "serv1escrow",
//         "to_address": task.Claimer,
//         "amount": []map[string]string{
//             {
//                 "denom": "microSERVDR",
//                 "amount": task.Bounty,
//             },
//         },
//     }

//     metadata := map[string]interface{}{
//         "type": "approve_task",
//         "task_id": task.ID,
//         "status": TaskStatusApproved,
//     }
//     metadataBytes, err := json.Marshal(metadata)
//     if err != nil {
//         return fmt.Errorf("failed to marshal metadata: %v", err)
//     }

//     msgBytes, err := json.Marshal(msg)
//     if err != nil {
//         return fmt.Errorf("failed to marshal message: %v", err)
//     }

//     tx := Transaction{
//         Body: TxBody{
//             Messages: []json.RawMessage{msgBytes},
//             Memo: string(metadataBytes),
//             TimeoutHeight: uint64(time.Now().Unix() + 300),
//         },
//         AuthInfo: AuthInfo{
//             SignerInfos: []SignerInfo{
//                 {
//                     PublicKey: PublicKey{
//                         Type: "/cosmos.crypto.secp256k1.PubKey",
//                         Key: c.walletKeys[approver].PubKey,
//                     },
//                     ModeInfo: ModeInfo{
//                         Single: Single{
//                             Mode: "SIGN_MODE_DIRECT",
//                         },
//                     },
//                     Sequence: "0",
//                 },
//             },
//             Fee: Fee{
//                 Amount: []Coin{
//                     {
//                         Denom: "microSERVDR",
//                         Amount: "1000",
//                     },
//                 },
//                 GasLimit: 200000,
//             },
//         },
//         Signatures: make([]string, 0),
//     }

//     txBytes, err := json.Marshal(tx)
//     if err != nil {
//         return fmt.Errorf("failed to marshal transaction: %v", err)
//     }
    
//     return c.submitTransaction(txBytes)
// }

// func (c *BlockchainClient) submitTransaction(txBytes []byte) error {
//     // Create broadcast request
//     broadcastReq := map[string]interface{}{
//         "tx_bytes": base64.StdEncoding.EncodeToString(txBytes),
//         "mode":     "BROADCAST_MODE_SYNC",
//     }

//     reqBytes, err := json.Marshal(broadcastReq)
//     if err != nil {
//         return fmt.Errorf("failed to marshal broadcast request: %v", err)
//     }

//     url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", c.restEndpoint)
//     log.Printf("Submitting transaction to: %s", url)
//     log.Printf("Transaction payload: %s", string(reqBytes))

//     resp, err := http.Post(
//         url,
//         "application/json",
//         bytes.NewBuffer(reqBytes),
//     )
//     if err != nil {
//         return fmt.Errorf("failed to submit transaction: %v", err)
//     }
//     defer resp.Body.Close()

//     body, err := ioutil.ReadAll(resp.Body)
//     if err != nil {
//         return fmt.Errorf("failed to read response: %v", err)
//     }

//     log.Printf("Response status: %d", resp.StatusCode)
//     log.Printf("Response body: %s", string(body))

//     var response struct {
//         TxResponse struct {
//             Code      int    `json:"code"`
//             Data      string `json:"data"`
//             RawLog    string `json:"raw_log"`
//             TxHash    string `json:"txhash"`
//             GasWanted string `json:"gas_wanted"`
//             GasUsed   string `json:"gas_used"`
//         } `json:"tx_response"`
//     }

//     if err := json.Unmarshal(body, &response); err != nil {
//         return fmt.Errorf("failed to parse response: %v", err)
//     }

//     if response.TxResponse.Code != 0 {
//         return fmt.Errorf(response.TxResponse.RawLog)
//     }

//     return nil
// }

// func (c *BlockchainClient) ClaimTask(taskID string, claimer string, proof string) error {
//     msg := map[string]interface{}{
//         "@type":       "/cosmos.bank.v1beta1.MsgSend",
//         "from_address": claimer,
//         "proof":       proof,
//         "status":      TaskStatusClaimed,
//     }

//     msgBytes, err := json.Marshal(msg)
//     if err != nil {
//         return fmt.Errorf("failed to marshal message: %v", err)
//     }

//     tx := Transaction{
//         Body: TxBody{
//             Messages:      []json.RawMessage{msgBytes},
//             Memo:         fmt.Sprintf("claim-%s", taskID),
//             TimeoutHeight: uint64(time.Now().Unix() + 300),
//         },
//         AuthInfo: AuthInfo{
//             SignerInfos: []SignerInfo{
//                 {
//                     PublicKey: PublicKey{
//                         Type: "/cosmos.crypto.secp256k1.PubKey",
//                         Key:  c.walletKeys[claimer].PubKey,
//                     },
//                     ModeInfo: ModeInfo{
//                         Single: Single{
//                             Mode: "SIGN_MODE_DIRECT",
//                         },
//                     },
//                     Sequence: "0",
//                 },
//             },
//             Fee: Fee{
//                 Amount: []Coin{
//                     {
//                         Denom:  "microSERVDR",
//                         Amount: "1000",
//                     },
//                 },
//                 GasLimit: 200000,
//             },
//         },
//         Signatures: make([]string, 0),
//     }

//     txBytes, err := json.Marshal(tx)
//     if err != nil {
//         return fmt.Errorf("failed to marshal transaction: %v", err)
//     }
    
//     return c.submitTransaction(txBytes)
// }

// // Improved transaction signing method

// func (c *BlockchainClient) signTransaction(tx *Transaction, signer string) error {
//     wallet, exists := c.walletKeys[signer]
//     if !exists {
//         return fmt.Errorf("wallet key not found for %s", signer)
//     }

//     // In a production environment, you would use a secure way to handle private keys
//     // For this example, we'll create a deterministic key from the private key string
//     // WARNING: This is for development only and is NOT secure for production!
    
//     // Create a SHA-256 hash of the private key string to get consistent bytes
//     hasher := sha256.New()
//     hasher.Write([]byte(wallet.PrivKey))
//     privKeyBytes := hasher.Sum(nil)
    
//     // Create the private key
//     privKey := secp256k1.PrivKey{Key: privKeyBytes}

//     // Sign documents need to be created according to the Cosmos SDK standards
//     // For SIGN_MODE_DIRECT, we need to sign the serialized transaction
    
//     // For a real implementation, you would use the proper proto encoding
//     // This is a simplified example
//     signBytes, err := json.Marshal(struct {
//         Body       TxBody   `json:"body"`
//         AuthInfo   AuthInfo `json:"auth_info"`
//         ChainID    string   `json:"chain_id"`
//     }{
//         Body:     tx.Body,
//         AuthInfo: tx.AuthInfo,
//         ChainID:  c.chainID,
//     })
    
//     if err != nil {
//         return fmt.Errorf("failed to marshal sign docs: %v", err)
//     }

//     // Create signature
//     signature, err := privKey.Sign(signBytes)
//     if err != nil {
//         return fmt.Errorf("failed to sign transaction: %v", err)
//     }

//     // Add signature to transaction
//     tx.Signatures = append(tx.Signatures, base64.StdEncoding.EncodeToString(signature))

//     return nil
// }


// // Admin management functions
// func (c *BlockchainClient) IsAdmin(address string) bool {
//     return c.adminWallets[address]
// }

// func (c *BlockchainClient) GetTestWallets() []string {
//     return []string{
//         "cosmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd80udgsy", // Valid Cosmos SDK address
//         "cosmos1xyxs3skf3f4jfqeuv89yyaqvjc6lffavxqhc8g",  // Valid Cosmos SDK address
//     }
// }


// func (c *BlockchainClient) GetChainID() string {
//     return c.chainID
// }

// func (c *BlockchainClient) GetRPCEndpoint() string {
//     return c.rpcEndpoint
// }

// func (c *BlockchainClient) GetRESTEndpoint() string {
//     return c.restEndpoint
// }

// // Improved implementation for the ListTasks method in BlockchainClient

// func (c *BlockchainClient) ListTasks() ([]intTypes.Task, error) {
//     // In a real blockchain implementation, we need to query transactions
//     // with our task metadata in the memo field
//     url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", c.restEndpoint)
    
//     // For a real implementation, we would need pagination and filtering
//     // Here we're keeping it simple but would need to be expanded for production
//     requestBody := map[string]interface{}{
//         "pagination": map[string]interface{}{
//             "limit": "100",  // Limit to latest 100 transactions
//         },
//         // You might add custom events or filters specific to your chain
//         "events": []string{
//             "message.action='/cosmos.bank.v1beta1.MsgSend'",
//             // You could add other filters here
//         },
//     }
    
//     requestBytes, err := json.Marshal(requestBody)
//     if err != nil {
//         return nil, fmt.Errorf("failed to marshal request: %v", err)
//     }
    
//     // Make the HTTP POST request to get transactions
//     resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBytes))
//     if err != nil {
//         return nil, fmt.Errorf("failed to fetch transactions: %v", err)
//     }
//     defer resp.Body.Close()
    
//     if resp.StatusCode != http.StatusOK {
//         body, _ := ioutil.ReadAll(resp.Body)
//         return nil, fmt.Errorf("failed to query transactions, status: %d, response: %s", 
//             resp.StatusCode, string(body))
//     }
    
//     // Parse the response
//     body, err := ioutil.ReadAll(resp.Body)
//     if err != nil {
//         return nil, fmt.Errorf("failed to read response: %v", err)
//     }
    
//     var response struct {
//         TxResponses []struct {
//             TxHash    string          `json:"txhash"`
//             Height    string          `json:"height"`
//             Timestamp string          `json:"timestamp"`
//             Code      int             `json:"code"`
//             RawLog    string          `json:"raw_log"`
//             Tx        json.RawMessage `json:"tx"`
//         } `json:"tx_responses"`
//     }
    
//     if err := json.Unmarshal(body, &response); err != nil {
//         return nil, fmt.Errorf("failed to parse response: %v", err)
//     }
    
//     // Process transactions to extract tasks
//     var tasks []intTypes.Task
    
//     for _, txResp := range response.TxResponses {
//         // Skip failed transactions
//         if txResp.Code != 0 {
//             continue
//         }
        
//         // Parse the transaction
//         var tx Transaction
//         if err := json.Unmarshal(txResp.Tx, &tx); err != nil {
//             log.Printf("Warning: Failed to parse transaction %s: %v", txResp.TxHash, err)
//             continue
//         }
        
//         // Check if the memo contains task metadata
//         if tx.Body.Memo == "" {
//             continue
//         }
        
//         // Try to parse the memo as task metadata
//         var metadata map[string]interface{}
//         if err := json.Unmarshal([]byte(tx.Body.Memo), &metadata); err != nil {
//             // Not a task metadata, skip
//             continue
//         }
        
//         // Check if it's a task-related transaction
//         taskType, ok := metadata["type"].(string)
//         if !ok {
//             continue
//         }
        
//         // Process based on task type
//         switch taskType {
//         case "create_task":
//             taskID, _ := metadata["task_id"].(string)
//             title, _ := metadata["title"].(string)
//             description, _ := metadata["description"].(string)
//             status, _ := metadata["status"].(string)
            
//             // Extract the sender and amount from the message
//             if len(tx.Body.Messages) == 0 {
//                 continue
//             }
            
//             var msg map[string]interface{}
//             if err := json.Unmarshal(tx.Body.Messages[0], &msg); err != nil {
//                 log.Printf("Warning: Failed to parse message in tx %s: %v", txResp.TxHash, err)
//                 continue
//             }
            
//             creator, _ := msg["from_address"].(string)
            
//             // Get the bounty amount
//             bounty := "0"
//             if amounts, ok := msg["amount"].([]interface{}); ok && len(amounts) > 0 {
//                 if amount, ok := amounts[0].(map[string]interface{}); ok {
//                     bounty, _ = amount["amount"].(string)
//                 }
//             }
            
//             task := intTypes.Task{
//                 ID:          taskID,
//                 Title:       title,
//                 Description: description,
//                 Creator:     creator,
//                 Bounty:      bounty,
//                 Status:      status,
//             }
            
//             tasks = append(tasks, task)
            
//         case "claim_task":
//             taskID, _ := metadata["task_id"].(string)
            
//             // Find the existing task
//             var existingTask *intTypes.Task
//             for i, t := range tasks {
//                 if t.ID == taskID {
//                     existingTask = &tasks[i]
//                     break
//                 }
//             }
            
//             if existingTask == nil {
//                 // Task not found, could be a task created before our query window
//                 continue
//             }
            
//             // Update the task status
//             existingTask.Status = "CLAIMED"
            
//             // Extract claimer from message
//             if len(tx.Body.Messages) > 0 {
//                 var msg map[string]interface{}
//                 if err := json.Unmarshal(tx.Body.Messages[0], &msg); err != nil {
//                     continue
//                 }
                
//                 claimer, _ := msg["from_address"].(string)
//                 proof, _ := msg["proof"].(string)
                
//                 existingTask.Claimer = claimer
//                 existingTask.Proof = proof
//             }
            
//         case "approve_task":
//             taskID, _ := metadata["task_id"].(string)
            
//             // Find the existing task
//             var existingTask *intTypes.Task
//             for i, t := range tasks {
//                 if t.ID == taskID {
//                     existingTask = &tasks[i]
//                     break
//                 }
//             }
            
//             if existingTask == nil {
//                 // Task not found, could be a task created before our query window
//                 continue
//             }
            
//             // Update the task status
//             existingTask.Status = "APPROVED"
//         }
//     }
    
//     // If we couldn't find any tasks from the blockchain, return mock data for testing
//     if len(tasks) == 0 {
//         log.Printf("Warning: No tasks found on blockchain, returning mock tasks for testing")
//         tasks = []intTypes.Task{
//             {
//                 ID:          "task-1",
//                 Title:       "Test Task 1",
//                 Description: "This is a test task from blockchain",
//                 Creator:     "serv1ph3trl8q0mtc6dfz5xsqq6pksl7w7rxx0hk3q",
//                 Bounty:      "1000",
//                 Status:      "OPEN",
//             },
//             {
//                 ID:          "task-2",
//                 Title:       "Test Task 2",
//                 Description: "Another test task from blockchain",
//                 Creator:     "serv1mj9k8h7l6n5m4p3q2r1s0t9v8w7x6y5z4a3b",
//                 Bounty:      "2000",
//                 Status:      "CLAIMED",
//                 Claimer:     "serv1ph3trl8q0mtc6dfz5xsqq6pksl7w7rxx0hk3q",
//                 Proof:       "Test proof of work",
//             },
//         }
//     }
    
//     return tasks, nil
// }

// func (c *BlockchainClient) ValidateAddress(address string) bool {
//     // Check if the address is empty
//     if address == "" {
//         return false
//     }
    
//     // Check if the address has the correct prefix - allow both serv1 format and cosmos1 format for testing
//     if !strings.HasPrefix(address, "serv1") && !strings.HasPrefix(address, "cosmos1") {
//         log.Printf("Address validation failed for %s: incorrect prefix", address)
//         return false
//     }
    
//     // Try to decode the address
//     _, err := sdk.AccAddressFromBech32(address)
//     if err != nil {
//         log.Printf("Address validation failed for %s: %v", address, err)
//         return false
//     }
    
//     return true
// }

// Token distribution related functions
type TokenDistribution struct {
    FromAddress string
    ToAddress   string
    Amount      string
    Denom       string
    TxHash      string
}

func (c *BlockchainClient) DistributeTokens(task intTypes.Task) error {
    // Mock version
    log.Printf("Mock: Distributing %s microSERVDR tokens to %s for task %s", 
        task.Bounty, task.Claimer, task.ID)
    return nil

    /* Real blockchain version (commented out)
    // Create token transfer message
    msg := map[string]interface{}{
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": c.GetEscrowAddress(),  // Task bounty is held in escrow
        "to_address": task.Claimer,
        "amount": []map[string]string{
            {
                "denom": "microSERVDR",
                "amount": task.Bounty,
            },
        },
    }

    // Add distribution metadata
    metadata := map[string]interface{}{
        "type": "token_distribution",
        "task_id": task.ID,
        "distribution_type": "task_completion",
    }

    metadataBytes, err := json.Marshal(metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %v", err)
    }

    // Create and sign transaction
    tx := Transaction{
        Body: TxBody{
            Messages: []json.RawMessage{msgBytes},
            Memo: string(metadataBytes),
            TimeoutHeight: uint64(time.Now().Unix() + 300),
        },
        AuthInfo: AuthInfo{
            SignerInfos: []SignerInfo{
                {
                    PublicKey: PublicKey{
                        Type: "/cosmos.crypto.secp256k1.PubKey",
                        Key: c.walletKeys[c.GetEscrowAddress()].PubKey,
                    },
                    ModeInfo: ModeInfo{
                        Single: Single{
                            Mode: "SIGN_MODE_DIRECT",
                        },
                    },
                    Sequence: "0",
                },
            },
            Fee: Fee{
                Amount: []Coin{
                    {
                        Denom: "microSERVDR",
                        Amount: "1000",
                    },
                },
                GasLimit: 200000,
            },
        },
        Signatures: make([]string, 0),
    }

    return c.submitTransaction(txBytes)
    */
}

func (c *BlockchainClient) GetEscrowAddress() string {
    // For mock mode, return a constant address
    return "serv1escrow000000000000000000000000000000"

    /* Real blockchain version (commented out)
    // Generate deterministic escrow address
    hasher := sha256.New()
    hasher.Write([]byte("escrow"))
    hash := hasher.Sum(nil)
    addr := sdk.AccAddress(hash[:20])
    return addr.String()
    */
}

func (c *BlockchainClient) LockTaskBounty(task intTypes.Task) error {
    // Mock version
    log.Printf("Mock: Locking %s microSERVDR tokens from %s in escrow for task %s", 
        task.Bounty, task.Creator, task.ID)
    return nil

    /* Real blockchain version (commented out)
    // Transfer tokens to escrow
    msg := map[string]interface{}{
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": task.Creator,
        "to_address": c.GetEscrowAddress(),
        "amount": []map[string]string{
            {
                "denom": "microSERVDR",
                "amount": task.Bounty,
            },
        },
    }

    metadata := map[string]interface{}{
        "type": "lock_bounty",
        "task_id": task.ID,
    }

    // Create and submit transaction...
    */
}

func (c *BlockchainClient) GetTokenBalance(address string) (string, error) {
    // Mock version
    return "1000000", nil

    /* Real blockchain version (commented out)
    url := fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", c.restEndpoint, address)
    resp, err := http.Get(url)
    if err != nil {
        return "", fmt.Errorf("failed to get balance: %v", err)
    }
    defer resp.Body.Close()

    var response struct {
        Balances []struct {
            Denom  string `json:"denom"`
            Amount string `json:"amount"`
        } `json:"balances"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", fmt.Errorf("failed to decode response: %v", err)
    }

    for _, balance := range response.Balances {
        if balance.Denom == "microSERVDR" {
            return balance.Amount, nil
        }
    }

    return "0", nil
    */
}

// Update ApproveTask to include token distribution
// func (c *BlockchainClient) ApproveTask(task intTypes.Task, approver string) error {
//     if !c.IsAdmin(approver) {
//         return fmt.Errorf("only admins can approve tasks")
//     }
    
//     existingTask, exists := c.tasks[task.ID]
//     if !exists {
//         return fmt.Errorf("task not found")
//     }
    
//     if existingTask.Status != "CLAIMED" {
//         return fmt.Errorf("task must be claimed before approval")
//     }

//     // Distribute tokens to claimer
//     if err := c.DistributeTokens(existingTask); err != nil {
//         return fmt.Errorf("failed to distribute tokens: %v", err)
//     }
    
//     existingTask.Status = "COMPLETED"
//     c.tasks[task.ID] = existingTask
    
//     log.Printf("Task %s approved by admin %s and tokens distributed", task.ID, approver)
//     return nil
// }

// // Update CreateTask to include bounty locking
// func (c *BlockchainClient) CreateTask(task intTypes.Task) error {
//     if task.ID == "" || task.Title == "" || task.Bounty == "" {
//         return fmt.Errorf("invalid task parameters")
//     }

//     // Lock bounty in escrow
//     if err := c.LockTaskBounty(task); err != nil {
//         return fmt.Errorf("failed to lock bounty: %v", err)
//     }
    
//     // Store task in memory
//     c.tasks[task.ID] = task
//     log.Printf("Created task: %+v and locked bounty", task)
//     return nil
// }