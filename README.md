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
│   └── utils/              # Checker utilities
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

## VPS Deployment Guide

Deploying Downtime Tracker to an Ubuntu VPS is incredibly straightforward. This guide covers a production-ready setup utilizing `systemd` for background persistence and `Nginx` as a reverse proxy for SSL.

### 1. Initial VPS Setup
SSH into your Ubuntu server and install the core dependencies:
```bash
sudo apt update
sudo apt install -y golang-go redis-server nginx git
```

Install MongoDB (Follow MongoDB's [official Ubuntu installation guide](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/) as package signatures vary frequently).

### 2. Clone & Build
```bash
git clone https://github.com/CodemHax/downtimetracker.git
cd downtimetracker

# Build the production binary
go build -o tracker .
```

### 3. Environment Variables
Create your configuration file in the project folder:
```bash
nano .env
```
Ensure that `LINK` points to your public domain (e.g. `LINK=https://tracker.yourdomain.com`).

### 4. Setup Systemd Daemon
Create a system service so the app runs in the background and launches on startup automatically.

```bash
sudo nano /etc/systemd/system/downtimetracker.service
```

Add the following (replace `/path/to/downtimetracker` with your actual directory):
```ini
[Unit]
Description=Downtime Tracker Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/path/to/downtimetracker
ExecStart=/path/to/downtimetracker/tracker
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable downtimetracker
sudo systemctl start downtimetracker
```

### 5. Setup Nginx Reverse Proxy
To securely expose your app using a domain name, configure Nginx to route traffic to the Go app running locally on port `8080`.

```bash
sudo nano /etc/nginx/sites-available/downtimetracker
```

```nginx
server {
    listen 80;
    server_name tracker.yourdomain.com; # Replace with your domain

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable the site and restart Nginx:
```bash
sudo ln -s /etc/nginx/sites-available/downtimetracker /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 6. SSL Configuration (Certbot / HTTPS)
Finally, encrypt your connection:
```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d tracker.yourdomain.com
```

Your downtime tracker is now securely deployed and globally accessible!

## License

MIT
