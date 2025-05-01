# Fyne Chat Application

This is a simple chat application built using the Fyne framework in Go. The application allows users to log in and participate in a chat room where they can send and receive messages in real-time.

## Project Structure

```
fyne-chat-app
├── cmd
│   └── app
│       └── main.go         # Entry point of the application
├── internal
│   ├── ui
│   │   ├── chatview.go     # Chat interface
│   │   └── loginview.go    # Login interface
│   └── network
│       └── client.go       # Network client for chat server
├── go.mod                   # Module definition and dependencies
└── README.md                # Project documentation
```

## Features

- User authentication through a login interface.
- Real-time messaging in a chat room.
- Simple and intuitive user interface.

## Setup Instructions

1. Clone the repository:
   ```
   git clone <repository-url>
   cd fyne-chat-app
   ```

2. Install the necessary dependencies:
   ```
   go mod tidy
   ```

3. Run the application:
   ```
   go run cmd/app/main.go
   ```

## Usage

- Upon starting the application, users will be presented with a login screen.
- After successful authentication, users will be redirected to the chat interface where they can send and receive messages.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.