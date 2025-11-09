# garmin-to-ido

A Golang CLI tool to synchronize bike activities from Garmin Connect to iDO Sport.

## Features

- Sync bike activities from Garmin Connect to iDO Sport
- By default, syncs today and yesterday's activities
- Specify a custom date to sync
- Uses Garmin Connect API with username/password authentication
- Uses browser automation (chromedp) for iDO Sport (no official API available)

## Prerequisites

- Go 1.21 or higher
- Chrome/Chromium browser (for headless browser automation)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd garmin-to-ido
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o garmin-to-ido
```

## Configuration

1. Copy the example configuration file:
```bash
cp .env.example .env
```

2. Edit `.env` and add your credentials:
```
GARMIN_USERNAME=your-garmin-email@example.com
GARMIN_PASSWORD=your-garmin-password
IDO_USERNAME=your-ido-email@example.com
IDO_PASSWORD=your-ido-password
```

**Note:** Keep your `.env` file secure and never commit it to version control.

## Usage

### Sync today and yesterday (default behavior)
```bash
./garmin-to-ido
```

### Sync a specific date
```bash
./garmin-to-ido -date 2025-01-15
```

### Use a custom config file
```bash
./garmin-to-ido -config /path/to/config.env
```

### All options
```bash
./garmin-to-ido -h
```

## Project Structure

```
garmin-to-ido/
├── main.go                      # Entry point and CLI parsing
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── garmin/
│   │   └── client.go           # Garmin Connect API client
│   ├── ido/
│   │   └── client.go           # iDO Sport browser automation
│   └── sync/
│       └── syncer.go           # Synchronization logic
├── .env                         # Configuration file (not in git)
├── .env.example                 # Example configuration
└── README.md
```

## TODO / Known Issues

- **iDO Sport upload implementation is incomplete**: The `internal/ido/client.go` file contains a placeholder for the upload functionality. You need to:
  1. Inspect the iDO Sport website to understand their upload mechanism
  2. Implement the actual upload logic using chromedp
  3. Handle file uploads (GPX format from Garmin)
  4. Add error handling for upload failures

## Development

### Run without building
```bash
go run main.go
```

### Run tests
```bash
go test ./...
```

### Build for different platforms
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o garmin-to-ido-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o garmin-to-ido-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o garmin-to-ido.exe
```

## How it works

1. **Authentication**:
   - Logs into Garmin Connect using official API endpoints
   - Logs into iDO Sport using browser automation

2. **Activity Retrieval**:
   - Fetches activities from Garmin for the specified date(s)
   - Filters for cycling/bike activities only

3. **Synchronization**:
   - Downloads each activity in GPX format
   - Uploads to iDO Sport via browser automation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See LICENSE file for details.
