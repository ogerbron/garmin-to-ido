# Embedding the Python Script

This document explains how the Python script is embedded into the Go binary.

## How It Works

The project uses Go's `embed` package to include the Python script directly in the compiled binary. Here's what happens:

### 1. Embed Directive (`internal/garmin/embed.go`)

```go
package garmin

import _ "embed"

//go:embed garmin_client.py
var pythonScript []byte
```

The `//go:embed` directive tells the Go compiler to:
- Read the `garmin_client.py` file at compile time
- Store its contents in the `pythonScript` byte slice
- Include it in the final binary

### 2. Runtime Extraction (`internal/garmin/python_client.go`)

When the `Login()` method is called:
1. A temporary file is created using `os.CreateTemp()`
2. The embedded script bytes are written to this temp file
3. The temp file path is stored in `c.scriptPath`
4. Python commands are executed using this temp file

### 3. Cleanup

When `Logout()` is called, the temporary file is deleted.

## Benefits

1. **Single Binary Distribution**: Users only need to download one file
2. **No Installation Hassles**: No need to separately download the scripts folder
3. **Version Consistency**: The Python script version always matches the binary version

## Requirements

Users still need:
- Python 3.x installed
- `garminconnect` package: `pip install garminconnect`
- Chrome/Chromium browser (for iDO automation)

## File Structure

```
internal/garmin/
├── embed.go              # Contains the embed directive
├── garmin_client.py      # The Python script (copied from scripts/)
├── python_client.go      # Go wrapper that uses the embedded script
├── interface.go          # Garmin client interface
└── types.go             # Type definitions
```

## Development Notes

### Why Copy the Script?

The Python script exists in two locations:
- `scripts/garmin_client.py` - Original source (for development)
- `internal/garmin/garmin_client.py` - Copy for embedding

This is because `//go:embed` can only reference files in the same directory or subdirectories of the package. When making changes to the Python script, update `scripts/garmin_client.py` and copy it to `internal/garmin/`.

### Keeping Scripts in Sync

You can use this command to sync the script:

```bash
cp scripts/garmin_client.py internal/garmin/garmin_client.py
```

Or add it to your build process.

### Alternative: Use go:generate

You could automate this with a `//go:generate` directive in `embed.go`:

```go
//go:generate cp ../../scripts/garmin_client.py garmin_client.py
```

Then run `go generate ./...` before building.

## Testing

To verify the script is embedded:

```bash
# Build the binary
go build -o garmin-to-ido

# Check if the script content is in the binary
strings garmin-to-ido | grep "from garminconnect import"
```

You should see Python code in the output.

## Size Impact

The Python script adds approximately ~5KB to the binary size, which is negligible for a typical Go binary (usually several MB).
