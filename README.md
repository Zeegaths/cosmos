# Tokenized Task Bounty System

A decentralized task reward platform built with Go and Cosmos SDK that enables users to create, claim, and approve tasks with token bounties.

## Features

- Create tasks with token bounties
- Claim tasks with proof of work
- Admin approval system
- Token reward distribution
- Address generation and management

## Getting Started

### Prerequisites

- Go 1.19 or later
- Git

### Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/bounty-system.git
cd bounty-system
```

2. Install dependencies
```bash
go mod tidy
```

3. Run the server
```bash
go run cmd/api/main.go
```

## API Usage

### 1. Generate an Address

```bash
# Generate a new address
curl -X POST http://localhost:8080/generate-address \
-H "Content-Type: application/json" \
-d '{"seed": "user-1"}'
```

### 2. Create a Task

```bash
curl -X POST http://localhost:8080/tasks \
-H "Content-Type: application/json" \
-d '{
    "title": "Build Website",
    "description": "Create responsive React website",
    "bounty": "1000000",
    "creator": "serv1..."  # Use generated address
}'
```

### 3. List Tasks

```bash
curl http://localhost:8080/tasks
```

### 4. Claim a Task

```bash
curl -X PUT http://localhost:8080/tasks/{taskId}/claim \
-H "Content-Type: application/json" \
-d '{
    "claimer": "serv1...",  # Use generated address
    "proof": "https://github.com/proof"
}'
```

### 5. Approve Task (Admin)

```bash
curl -X PUT http://localhost:8080/admin/tasks/{taskId} \
-H "Content-Type: application/json" \
-H "X-Wallet-Address: {admin-address}"
```

## Project Structure

```
bounty-system/
├── cmd/
│   └── api/
│       └── main.go          # API server
├── internal/
│   ├── client/
│   │   └── blockchain.go    # Blockchain operations
│   └── types/
│       └── task.go          # Data structures
└── README.md
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/generate-address` | Generate new address |
| GET | `/addresses` | List all addresses |
| POST | `/tasks` | Create new task |
| GET | `/tasks` | List all tasks |
| PUT | `/tasks/{id}/claim` | Claim a task |
| PUT | `/admin/tasks/{id}` | Approve task (admin) |

## Task States

- `OPEN`: Task is available for claiming
- `CLAIMED`: Task has been claimed with proof
- `COMPLETED`: Task has been approved by admin

## Development Notes

Currently running in mock mode which:
- Generates valid addresses
- Simulates blockchain operations
- Maintains in-memory task state
- Validates admin operations

## Testing

Complete test flow:

1. Start server and note admin address:
```bash
go run cmd/api/main.go
```

2. Generate creator address:
```bash
curl -X POST http://localhost:8080/generate-address \
-H "Content-Type: application/json" \
-d '{"seed": "creator-1"}'
```

3. Create task using generated address:
```bash
curl -X POST http://localhost:8080/tasks \
-H "Content-Type: application/json" \
-d '{
    "title": "Test Task",
    "description": "Test Description",
    "bounty": "1000000",
    "creator": "GENERATED_ADDRESS"
}'
```

4. Generate claimer address:
```bash
curl -X POST http://localhost:8080/generate-address \
-H "Content-Type: application/json" \
-d '{"seed": "claimer-1"}'
```

5. Claim task:
```bash
curl -X PUT http://localhost:8080/tasks/TASK_ID/claim \
-H "Content-Type: application/json" \
-d '{
    "claimer": "CLAIMER_ADDRESS",
    "proof": "https://github.com/proof"
}'
```

6. Approve task using admin address:
```bash
curl -X PUT http://localhost:8080/admin/tasks/TASK_ID \
-H "Content-Type: application/json" \
-H "X-Wallet-Address: ADMIN_ADDRESS"
```

## Future Improvements

- Real blockchain integration
- Token management
- Multi-signature approvals
- Task categories
- Escrow contract implementation

## License

MIT License - see LICENSE file for details
