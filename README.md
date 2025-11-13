# i-chat-server
A real-time chat server built with Go. Enables instant messaging with WebSocket support, user authentication, and message persistence.

## Features
- Real-time messaging via WebSocket
- User authentication and session management
- Message history and persistence
- Room-based conversations
- Connection pooling and concurrent client handling

## Getting Started

### Prerequisites
- Go 1.19+
- PostgreSQL (or your database)

### Installation
```bash
git clone <repository-url>
cd i-chat-server
go mod download
```

### Running
```bash
go run main.go
```

The server will start on `localhost:8080` by default.

## Configuration
Create a `.env` file with:
```
PORT=8080
DATABASE_URL=postgres://user:pass@localhost/dbname
JWT_SECRET=your_secret_key
```

## API Endpoints
- `POST /auth/register` - Register new user
- `POST /auth/login` - User login
- `WS /chat` - WebSocket connection for messaging

## Contributing
Pull requests welcome. Please open an issue for major changes.

## License
MIT