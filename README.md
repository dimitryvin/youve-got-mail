# Mail Notification Project

Project that details how one would receive notifications for their mail getting delivered.

# Server

## Setup

1. Copy the environment file example:
```bash
cp .env.example .env
```

2. Edit the `.env` file with your configuration:
- `EMAIL_FROM`: Your sender email address (e.g., "Your Name <your-email@example.com>")
- `EMAIL_PASSWORD`: Your SMTP password
- `EMAIL_TO`: Comma-separated list of recipient email addresses
- `SMTP_HOST`: Your SMTP server hostname
- `SMTP_PORT`: Your SMTP server port (usually 587 for TLS)

## Running with Docker

```bash
docker-compose up -d
```

## Running without Docker

```bash
cd server
go run main.go
```

## Security Notes

1. Never commit the `.env` file to version control
2. Use strong passwords for SMTP authentication
3. Keep your environment variables secure
4. Use TLS for SMTP connections (port 587)

## API Endpoints

- `POST /mail-delivered`: Sends a notification that mail has been delivered

## License

MIT 
