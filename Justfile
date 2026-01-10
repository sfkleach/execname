default:
    @just --list

test: unittest lint fmt-check gosec tidy build

unittest:
    go test ./...

lint:
    echo "Running linter..."
    @if command -v ~/tools/ext/bin/golangci-lint >/dev/null 2>&1; then \
        ~/tools/ext/bin/golangci-lint run; \
    elif command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run; \
    else \
        echo "golangci-lint not found, falling back to go vet"; \
        echo "To install golangci-lint locally, run: just install-golangci-lint"; \
        go vet ./...; \
    fi

gosec:
    @echo "Running security scanner..."
    @if command -v ~/go/bin/gosec >/dev/null 2>&1; then \
        ~/go/bin/gosec -quiet -fmt=text ./...; \
    elif command -v gosec >/dev/null 2>&1; then \
        gosec -quiet -fmt=text ./...; \
    else \
        echo "gosec not found, skipping security scan"; \
        echo "To install gosec, run: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
    fi

# Install golangci-lint to ~/tools/ext/bin  
install-golangci-lint:
    @mkdir -p ~/tools/ext/bin
    GOBIN=~/tools/ext/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @echo "golangci-lint installed locally to this project in ~/tools/ext/bin/"
    @echo "Note that ~/tools/ext/bin is not assumed to be in your PATH"

fmt:
    go fmt ./...

# Check formatting without modifying files
fmt-check:
    ./tools/bin/go-fmt-check

tidy:
    go mod tidy

build:
    mkdir -p bin
    go build -ldflags "-X github.com/sfkleach/execman/pkg/version.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X github.com/sfkleach/execman/pkg/version.GitCommit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown') -X github.com/sfkleach/execman/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/execman ./cmd/execman

install:
    go install ./cmd/execman


# Initialize decision records
init-decisions:
    python3 scripts/decisions.py --init

# Add a new decision record
add-decision TOPIC:
    python3 scripts/decisions.py --add "{{TOPIC}}"
