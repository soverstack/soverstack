# Soverstack Launcher

Native Go launcher for Soverstack - A transparent proxy between users and the Docker-based Soverstack runtime.

## Architecture

```
User → launcher (native binary) → Docker container → CLI (Node.js) → Ansible/Terraform
```

## What This Does

The launcher is a **dumb proxy** with one job: orchestrate Docker container execution. It:

1. Reads `platform.yaml` to extract runtime version
2. Checks Docker is available
3. Pulls `soverstack/cli-runtime:<version>` image
4. Mounts current directory as `/workspace`
5. Forwards all arguments to the container
6. Streams stdin/stdout/stderr in real-time
7. Exits with the CLI's exit code

## What This Does NOT Do

- No business logic (validation, planning, generation)
- No command interpretation
- No file generation
- No secret management

**All intelligence lives in the CLI** (inside the Docker container).

## Installation

### macOS / Linux

```bash
brew install soverstack/tap/soverstack
```

### Windows

```powershell
scoop bucket add soverstack https://github.com/soverstack/scoop-bucket
scoop install soverstack
```

### Alternative: one-liner script (no package manager needed)

```bash
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/soverstack/cli-launcher/main/install.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/soverstack/cli-launcher/main/install.ps1 | iex
```

### Build from Source

Requires [Go 1.25+](https://go.dev/dl/).

```bash
git clone https://github.com/soverstack/cli-launcher.git
cd cli-launcher
make build
sudo mv soverstack /usr/local/bin/
```

### Verify Installation

```bash
soverstack --version
```

## Prerequisites

- **Docker**: Docker Desktop (Windows/macOS) or Docker Engine (Linux)
- **Internet**: Required for first-time image pull (subsequent runs use cached image)

## Usage

### Commands

| Command | Description |
|---------|-------------|
| `soverstack init [project-name]` | Initialize a new Soverstack project |
| `soverstack validate [path]` | Validate project structure and configuration |
| `soverstack plan [path]` | Show execution plan (desired vs current state) |
| `soverstack apply [path]` | Apply infrastructure changes |
| `soverstack add region [name]` | Add a new region to the project |
| `soverstack add zone [region] [zone-name]` | Add a new zone to a region |
| `soverstack generate ssh` | Generate or rotate SSH keys |

### Init Options

```bash
soverstack init my-infra
soverstack init my-infra --domain example.com --tier production
soverstack init my-infra --regions 'eu:paris,lyon;us:oregon' --non-interactive
```

| Option | Description |
|--------|-------------|
| `--domain <domain>` | Domain name (e.g., example.com) |
| `--tier <tier>` | Infrastructure tier: `local`, `production`, `enterprise` |
| `--regions <regions>` | Regions and zones (e.g., `eu:paris,lyon;us:oregon`) |
| `--non-interactive` | Skip interactive prompts |

### Validate / Plan / Apply Options

```bash
soverstack validate
soverstack plan --verbose
soverstack apply --debug
```

| Option | Description |
|--------|-------------|
| `-v, --verbose` | Show detailed output with field changes |
| `--debug` | Show debug information |

### Add Region Options

```bash
soverstack add region us
soverstack add region us --zones portland,seattle --generate-ssh-keys
```

| Option | Description |
|--------|-------------|
| `--zones <zones>` | Zones to create (comma-separated) |
| `--generate-ssh-keys` | Generate SSH keys for new datacenters |

### Add Zone Options

```bash
soverstack add zone eu paris
soverstack add zone eu paris --generate-ssh-keys
```

| Option | Description |
|--------|-------------|
| `--generate-ssh-keys` | Generate SSH keys for the new zone |

### Generate SSH Options

```bash
soverstack generate ssh              # Interactive mode
soverstack generate ssh --all        # All datacenters
soverstack generate ssh --region eu  # All DCs in a region
soverstack generate ssh --dc eu:zone-paris  # Specific DC
```

| Option | Description |
|--------|-------------|
| `--all` | Generate for all datacenters |
| `--region <region>` | Generate for all DCs in a region |
| `--dc <region:dc>` | Generate for a specific datacenter |

### How It Works

1. **User runs**: `soverstack plan`
2. **Launcher reads**: `platform.yaml` → `version: v1.0.0`
3. **Pulls image**: `docker pull ghcr.io/soverstack/cli-runtime:v1.0.0`
4. **Runs container**: mounts current directory at `/workspace`
5. **Forwards I/O**: stdin/stdout/stderr streamed in real-time
6. **Exits**: with the same exit code as the CLI

## Version Extraction

The launcher reads `platform.yaml` to determine which runtime version to use:

```yaml
name: prod
version: v1.0.0  # ← This field
domain: example.com
```

| Scenario | Behavior |
|----------|----------|
| `platform.yaml` exists with valid version | Use that version |
| `platform.yaml` missing or malformed | Default to `latest` |

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
Error: Failed to pull image soverstack/cli-runtime:v1.0.0
Suggestion: Check your internet connection and verify the version in platform.yaml
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

### Linux

- Requires Docker Engine
- User must be in `docker` group or run as root
- Container runs in `userns=host` mode to preserve file permissions

### macOS

- Uses Docker Desktop
- Native ARM64 support for M1/M2/M3/M4 chips

## Development

### Build

```bash
make build      # Build binary (version from VERSION file)
make test       # Run go vet + go test
make snapshot   # GoReleaser snapshot (all platforms)
make clean      # Remove build artifacts
```

### Versioning

The `VERSION` file is the single source of truth:

- **Dev**: `1.0.0-SNAPSHOT`
- **Release**: `1.0.0`

On push to `main`, CI reads `VERSION`, creates a git tag, and publishes the release.

### Project Structure

```
launcher/
├── VERSION                          # Version source of truth
├── Makefile                         # Build commands
├── go.mod                           # Go module definition
├── main.go                          # Entry point & orchestration
├── internal/
│   ├── docker/
│   │   ├── client.go               # Docker client init & availability check
│   │   ├── image.go                # Image pull with progress
│   │   └── container.go            # Container run & I/O forwarding
│   └── platform/
│       └── parser.go               # platform.yaml version extraction
├── install.sh                       # Linux/macOS install script
├── install.ps1                      # Windows install script
├── .goreleaser.yml                  # Release automation
└── .github/workflows/release.yml    # CI/CD
```

## Design Principles

1. **Launcher is a proxy** - No business logic, no validation, no generation
2. **Fail gracefully on config** - Default to `latest` if version extraction fails
3. **Forward everything** - All args passed through without interpretation
4. **Container is ephemeral** - AutoRemove ensures no state persistence
5. **I/O is transparent** - Attach before start, stream bidirectionally
6. **Exit codes matter** - Always preserve CLI's exit code

## License

AGPL-3.0
