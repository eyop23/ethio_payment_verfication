# Payment Verification API

A Go-based REST API for verifying payments from Ethiopian payment providers (CBE and TeleBirr).

## Features

- Verify CBE bank transfer receipts (PDF parsing)
- Verify TeleBirr mobile payment receipts (HTML parsing)
- Extract payment details: payer, receiver, amount, date, reference
- MongoDB storage for payments, users, and providers

## Setup

1. Copy environment file:
```bash
cp .env.example .env
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run main.go
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/user/ | Add a new user |
| GET | /api/user/ | Get all users |
| POST | /api/payment/providers | Add a payment provider |
| POST | /api/payment/verify | Verify a payment receipt |

## Tech Stack

- Go
- Gin (HTTP framework)
- MongoDB
- goquery (HTML parsing)
- ledongthuc/pdf (PDF parsing)
