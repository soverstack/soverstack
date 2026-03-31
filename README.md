# Soverstack Launcher

Native Go launcher for Soverstack - A transparent proxy between users and the Docker-based Soverstack runtime.

## Architecture

```
User → launcher (native binary) → Docker container → CLI (Node.js) → Ansible/Terraform
```

## What This Does

The launcher is a **dumb proxy** with one job: orchestrate Docker container execution. It:

1. ✅ Reads `platform.yaml` to extract runtime version
2. ✅ Checks Docker is available
3. ✅ Pulls `soverstack/runtime:<version>` image
4. ✅ Forwards all arguments to the container
5. ✅ Forwards all environment variables
6. ✅ Mounts current directory as `/workspace`
7. ✅ Streams stdin/stdout/stderr in real-time
8. ✅ Exits with the CLI's exit code

## What This Does NOT Do

- ❌ No business logic (validation, planning, generation)
- ❌ No command interpretation
- ❌ No file generation
- ❌ No secret management

**All intelligence lives in the CLI** (inside the Docker container).

## Installation

### Option 1: Download Pre-built Binary

Download the binary for your platform from the [Releases](https://github.com/soverstack/launcher/releases) page:

| Platform | Binary |
|----------|--------|
| Windows (x64) | `soverstack-windows-amd64.exe` |
| Linux (x64) | `soverstack-linux-amd64` |
| Linux (ARM64) | `soverstack-linux-arm64` |
| macOS (Intel) | `soverstack-darwin-amd64` |
| macOS (Apple Silicon) | `soverstack-darwin-arm64` |

**Windows:**
1. Download `soverstack-windows-amd64.exe`
2. Rename to `soverstack.exe`
3. Move to a directory in your PATH (e.g., `C:\Users\<you>\bin\`)
4. Or add its directory to your PATH environment variable

**Linux / macOS:**
```bash
# Download (replace OS and ARCH with your platform)
curl -L -o soverstack https://github.com/soverstack/launcher/releases/latest/download/soverstack-linux-amd64

# Make executable
chmod +x soverstack

# Move to PATH
sudo mv soverstack /usr/local/bin/
```

### Option 2: Build from Source

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone https://github.com/soverstack/launcher.git
cd launcher
go build -ldflags="-s -w" -o soverstack .

# Move to PATH
sudo mv soverstack /usr/local/bin/   # Linux/macOS
# or move soverstack.exe to a PATH directory on Windows
```

### Verify Installation

```bash
soverstack --version
# → soverstack launcher version v1.0.0
```

## Prerequisites

- **Docker**: Docker Desktop (Windows/macOS) or Docker Engine (Linux)
- **Internet**: Required for first-time image pull (subsequent runs use cached image)

## Building from Source

### Quick Build (Current Platform)

```bash
go build -o soverstack .
```

### Cross-Platform Build

```bash
# Build for all platforms (Windows, Linux, macOS)
chmod +x build/build.sh
./build/build.sh v1.0.0

# Output:
# dist/soverstack-windows-amd64.exe
# dist/soverstack-linux-amd64
# dist/soverstack-linux-arm64
# dist/soverstack-darwin-amd64
# dist/soverstack-darwin-arm64
```

### Build with Version

```bash
go build -ldflags="-X main.Version=v1.0.0" -o soverstack .
```

## Usage

### Basic Commands

```bash
# Validate platform configuration
soverstack validate platform.yaml

# Generate execution plan
soverstack plan platform.yaml

# Apply infrastructure changes
soverstack apply

# Update DNS nameservers
soverstack dns:update example.com

# Show version
soverstack --version

# Show help
soverstack --help
```

### How It Works

1. **User runs**: `soverstack plan platform.yaml`
2. **Launcher captures**: `args = ["plan", "platform.yaml"]`
3. **Extracts version**: Reads `platform.yaml` → `version: v1.0.0`
4. **Pulls image**: `docker pull soverstack/runtime:v1.0.0`
5. **Captures env vars**: All environment variables from host
6. **Runs container**:
   ```bash
   docker run --rm \
     -v $PWD:/workspace \
     -w /workspace \
     -e ENV_VAR1=value1 \
     -e ENV_VAR2=value2 \
     ... \
     soverstack/runtime:v1.0.0 \
     plan platform.yaml
   ```
7. **Forwards I/O**: stdin/stdout/stderr streamed in real-time
8. **Exits**: With the same exit code as the CLI

## Environment Variables

**All environment variables** are automatically forwarded to the container, including:

- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` - For S3 state storage
- `STATE_BUCKET` - S3 bucket for state (default: `soverstack-state`)
- `NAMECHEAP_USER`, `NAMECHEAP_KEY`, `NAMECHEAP_ACCOUNT`, `NAMECHEAP_IP` - For DNS operations
- `GODADDY_KEY`, `GODADDY_SECRET` - For DNS operations
- Any other environment variables

No filtering, no selection - everything is forwarded.

## Version Extraction

The launcher reads `platform.yaml` to determine which runtime version to use:

```yaml
name: prod
version: v1.0.0  # ← This field
domain: example.com
# ... rest of config
```

### Version Resolution Logic

| Scenario | Behavior |
|----------|----------|
| `platform.yaml` exists with valid version | Use that version |
| `platform.yaml` missing | Default to `latest` |
| `platform.yaml` malformed | Default to `latest` |
| Version field empty | Default to `latest` |

**Graceful degradation** - the launcher never fails on version extraction.

## Error Handling

### Docker Not Available

**Windows/macOS**:
```
Error: Docker is not available: Cannot connect to Docker daemon
Suggestion: Start Docker Desktop
```

**Linux**:
```
Error: Docker is not available: Cannot connect to Docker daemon
Suggestion:
  - Check if Docker daemon is running: sudo systemctl status docker
  - Start Docker: sudo systemctl start docker
  - Ensure your user is in the 'docker' group: sudo usermod -aG docker $USER
```

### Image Pull Failed

```
Error: Failed to pull image soverstack/runtime:v1.0.0
Suggestion: Check your internet connection and verify the version in platform.yaml
```

### CLI Error (Non-Zero Exit Code)

The launcher exits with the **same exit code** as the CLI:

```bash
# CLI fails with exit code 1
soverstack validate broken.yaml
echo $?  # Returns 1
```

## Signal Handling

**Ctrl+C** (SIGINT) is handled gracefully:

1. User presses Ctrl+C
2. Launcher catches signal
3. Container is stopped (10 second timeout)
4. Launcher exits

## Platform-Specific Behavior

### Windows

- Uses Docker Desktop
- Path conversion handled automatically by Docker Desktop
- No special file permission handling needed

### Linux

- Requires Docker Engine
- User must be in `docker` group or run as root
- Container runs in `userns=host` mode to preserve file permissions
- Files created by container have same ownership as host user

### macOS

- Uses Docker Desktop
- Native ARM64 support for M1/M2/M3 chips
- Path and permission handling automatic

## Project Structure

```
launcher/
├── go.mod                           # Go module definition
├── main.go                          # Entry point & orchestration
├── internal/
│   ├── docker/
│   │   ├── client.go               # Docker client init & availability check
│   │   ├── image.go                # Image pull with progress
│   │   └── container.go            # Container run & I/O forwarding
│   ├── platform/
│   │   └── parser.go               # Platform.yaml version extraction
│   └── environment/
│       └── env.go                  # Environment variable capture
└── build/
    └── build.sh                    # Cross-compilation script
```

## Testing

### Manual Testing

```bash
# Build
go build -o soverstack .

# Test version extraction
echo "version: v1.0.0" > platform.yaml
./soverstack validate platform.yaml

# Test with missing platform.yaml (should use "latest")
rm platform.yaml
./soverstack --version

# Test environment forwarding
export TEST_VAR=hello
./soverstack validate platform.yaml
# (Check in container that TEST_VAR is available)

# Test Ctrl+C handling
./soverstack plan platform.yaml
# Press Ctrl+C - should stop gracefully
```

### Unit Tests

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/platform/...
```

## Binary Size

Expected binary sizes (with `-ldflags="-s -w"`):

- **Uncompressed**: ~15-20 MB
- **UPX compressed**: ~5-7 MB (optional)

## Performance

- **Cold start** (first run, image pull): 3-5 seconds
- **Warm start** (cached image): < 1 second
- **Memory footprint**: < 50 MB

## Security Considerations

### Environment Variables

- All env vars are forwarded (including secrets)
- **Mitigation**: Container is ephemeral (`AutoRemove: true`)
- No logging of env vars
- No persistence of container state

### Volume Mounting

- Only current working directory is mounted (not entire filesystem)
- Runtime container is trusted (official Soverstack image)
- Linux: User permissions preserved via `userns=host`

### Container Cleanup

- Containers are automatically removed on exit (`AutoRemove: true`)
- No accumulation of stopped containers
- No state persistence

## Troubleshooting

### Binary won't execute (Linux/macOS)

```bash
chmod +x soverstack
```

### "permission denied" on Linux

Ensure your user is in the docker group:

```bash
sudo usermod -aG docker $USER
# Log out and back in for changes to take effect
```

### Windows: "Docker daemon not responding"

Start Docker Desktop and wait for it to fully initialize.

### Image pull hangs

Check your internet connection. First-time pull requires downloading the full runtime image.

## Design Principles

From the [Architecture README](../README.md):

1. **Launcher is a proxy** - No business logic, no validation, no generation
2. **Fail gracefully on config** - Default to "latest" if version extraction fails
3. **Forward everything** - All args and env vars passed through without interpretation
4. **Container is ephemeral** - AutoRemove ensures no state persistence
5. **I/O is transparent** - Attach before start, stream bidirectionally
6. **Exit codes matter** - Always preserve CLI's exit code

## Contributing

### Code Style

- Run `go fmt` before committing
- Follow standard Go conventions
- Keep functions small and focused
- Document exported functions

### Adding Features

**Think twice before adding features to the launcher.**

The launcher's power comes from its simplicity. If you're tempted to add:

- Command parsing → Put it in the CLI
- Validation logic → Put it in the CLI
- File generation → Put it in the CLI
- Configuration management → Put it in the CLI

The launcher should stay **dumb**.

## License

Proprietary - Soverstack

## Links

- **Main README**: [../README.md](../README.md)
- **Infrastructure Docs**: [../readmeInfra.md](../readmeInfra.md)
- **CLI Source**: [../runtine/](../runtine/)
