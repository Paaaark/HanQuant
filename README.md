# HanQuant

## Project Structure

```
HanQuant/
│
├── cmd/ # Entry points (main packages)
│ └── server/ # e.g., main.go for running API server
│ └── main.go
│
├── internal/ # Internal app-specific code (not for public use)
│ ├── config/ # Viper-based config loading
│ ├── server/ # HTTP & WebSocket server setup
│ ├── handler/ # REST API handlers (e.g., /prices/:symbol)
│ ├── websocket/ # WebSocket server and provider client
│ ├── data/ # Logic for fetching & caching financial data
│ └── strategy/ # (Milestone 4) Algorithm engine + backtesting
│
├── pkg/ # Public reusable packages (optional)
│ └── utils/ # Utility functions (e.g., time parsing, math)
│
├── api/ # API types, requests/responses, DTOs
│ └── models.go
│
├── scripts/ # Scripts (migrations, data seeding, tools)
│
├── configs/ # YAML configuration files
│ └── config.yaml
│
├── .env # Environment variables (for viper)
├── go.mod
├── go.sum
└── README.md
```
