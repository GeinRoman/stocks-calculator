## Running the Project

Using Docker Compose

```
docker compose up --build
```

Running CLI
```
go install cmd/cli/stcalc.go
```
then use stcalc in your terminal

# Stock Portfolio Rebalancer

A Go-based stock portfolio management and rebalancing system with a CLI client and backend service.

The project helps maintain target portfolio allocations across stock groups and individual stocks. It calculates how to rebalance a portfolio by suggesting which stocks to buy or sell while preserving predefined ratios.

## Features

* Portfolio management with multiple user profiles
* Group-based portfolio organization (e.g. Tech, Energy, Finance)
* Adjustable weight ratios between groups
* Portfolio rebalancing calculations
* CLI application for user interaction
* Backend API service
* Database-backed persistence
* Multi-device access through account system
* Docker support
* Swagger API documentation

## How Rebalancing Works

The application uses weighted allocation rules:

1. Portfolio value is divided according to group weights
2. Each group allocation is distributed equally between stocks inside that group
3. The system compares target allocations with current holdings
4. Buy/sell recommendations are generated to rebalance the portfolio

Example:

Tech group: 60%
Energy group: 40%

If Tech contains 3 stocks, each stock receives 20% target allocation.

## Project Architecture

The project is divided into two independent parts:

1. CLI Application

The CLI is the main user interface.

Responsibilities:

* User authentication
* Managing profiles
* Managing groups
* Managing stocks
* Triggering rebalance operations
* Connecting to backend service

2. Backend Service

The backend stores portfolio data and provides stock-related functionality.

Responsibilities:

* User management
* Data persistence
* Authentication/token handling
* Portfolio calculations
* Stock price retrieval

This separation allows:

Multi-device synchronization
Centralized data storage
Multiple accounts
Future web/mobile clients

## Tech Stack

* Language: Go
* Database: PostgreSQL, SQL-based migrations (used with migrate golang tool)
* Authentication: Token-based JWT auth
* Containerization: Docker & Docker Compose
* Endpoints documentation: Swagger (generated with swaggo/swag golang tool)
* Testing tools: testify and mockery

## Project Structure
```
.
├── cmd/
│   ├── cli/                 # CLI entrypoint
│   └── server/              # Backend server entrypoint
│
├── internal/
│   ├── cli/                 # CLI implementation
│   ├── model/               # Shared models
│   └── server/              # Backend logic
│
├── migrations/              # Database migrations
├── docs/                    # Swagger documentation
├── mocks/                   # Test mocks
├── pkg/                     # Shared packages
├── compose.yaml             # Docker Compose configuration
├── Dockerfile
└── config.yaml
```

## Configuration

Server configuration is provided through:

config.yaml and .env

Default config:

config.yaml
```
max_connections: 100
jwt_key: dev_secret_key_do_not_use_in_production
jwt_expiration_time_min: 10
```

.env
```
DATABASE_USER=pguser
DATABASE_PASSWORD=1234
PORT=8989
```

## Swagger documentation is available in:
```
/docs
```
