# Downtime Tracker

A simple and scalable website monitoring service built with Go. It periodically checks if your websites are up and sends email alerts when they go down or recover.

## Features

- 🔍 **Website Monitoring** - Checks HTTPS websites every 10 minutes
- 📧 **Email Alerts** - Sends notifications when sites go down or recover
- 🔐 **Email Verification** - Users must verify their email before monitoring starts
- ⚡ **Smart Alerts** - Only sends emails when status changes (no spam!)
- 🚀 **Concurrent Checks** - Configurable worker pool for parallel checking
- 💾 **Status Tracking** - Redis-based status memory to prevent duplicate alerts

## Tech Stack

- **Go** - Backend API and monitoring service
- **Gin** - HTTP web framework
- **MongoDB** - User and website storage
- **Redis** - Status tracking and token storage
- **SMTP** - Email delivery

## Quick Start

### Prerequisites

- Go 1.21+
- MongoDB
- Redis
- SMTP server (e.g., smtp2go, SendGrid)

### Installation

```bash
git https://github.com/CodemHax/downtimetracker
cd downtimetracker
go mod download
```

### Configuration

Create a `.env` file:

```env
MONGODB_URI=mongodb://localhost:27017
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
SMTP_HOST=mail.smtp2go.com
SMTP_PORT=2525
SMTP_USER=your-username
SMTP_PASS=your-password
SMTP_FROM=noreply@yourdomain.com
LINK=http://localhost:8080
CHECKER_TIMEOUT=5000
CHECKER_CONCURRENCY=5
```

### Run

```bash
go run .
```

Server starts at `http://localhost:8080`

## API Endpoints

### Health Check
```
GET /ping
```
Response: `{"message": "pong"}`

### Add Website to Monitor
```
POST /add
Content-Type: application/json

{
    "email": "user@example.com",
    "website": "https://yoursite.com"
}
```

### Verify Email
```
GET /verify?email=user@example.com&token=YOUR_TOKEN
```

## API Documentation (Swagger)

After starting the server, access the Swagger UI at:

    http://localhost:8000/swagger/index.html

To update the docs after changing endpoints or comments, run:

    swag init

## Project Structure

```
downtimetracker/
├── main.go                 # Entry point
├── cmd/
│   └── api.go              # API handlers
├── internals/
│   ├── database/
│   │   ├── mongo/          # MongoDB client
│   │   └── redis/          # Redis client
│   ├── mail/               # SMTP email sender
│   ├── models/             # Data models
│   └── utlis/              # Checker utilities
├── .env                    # Configuration
├── .gitignore
└── README.md
```

## How It Works

1. User adds a website via `/add` endpoint
2. Verification email is sent to the user
3. User clicks the verification link
4. Service starts monitoring the website every 10 minutes
5. If site goes down → Email alert sent
6. If site recovers → Recovery email sent
7. No repeated emails while status unchanged

## License

MIT
