# PowerShell build/test helper for terraform-provider-registry
# Usage: .\make.ps1 <target>

param(
    [Parameter(Position=0)]
    [string]$Target = "build"
)

$BINARY = "terraform-provider-registry"
$HOSTNAME = "registry.terraform.io"
$NAMESPACE = "terraform-registry"
$NAME = "registry"
$VERSION = "0.1.0"
$OS_ARCH = "$(go env GOOS)_$(go env GOARCH)"
$INSTALL_DIR = "$env:USERPROFILE\.terraform.d\plugins\$HOSTNAME\$NAMESPACE\$NAME\$VERSION\$OS_ARCH"

switch ($Target) {
    "build" {
        go build -o "$BINARY.exe" .
    }
    "install" {
        go build -o "$BINARY.exe" .
        New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
        Move-Item -Force "$BINARY.exe" "$INSTALL_DIR\$BINARY.exe"
        Write-Host "Installed to $INSTALL_DIR"
    }
    "test" {
        go test ./... -v -count=1 -timeout 10m
    }
    "testacc" {
        if (-not $env:TF_REGISTRY_ENDPOINT) {
            Write-Error "TF_REGISTRY_ENDPOINT must be set"
            exit 1
        }
        $env:TF_ACC = "1"
        go test ./internal/provider/... -v -count=1 -timeout 120m
    }
    "fmt" {
        gofmt -s -w .
    }
    "lint" {
        golangci-lint run ./...
    }
    "docs" {
        go generate ./...
    }
    "tidy" {
        go mod tidy
    }
    "clean" {
        Remove-Item -Force -ErrorAction SilentlyContinue "$BINARY.exe"
    }
    default {
        Write-Error "Unknown target: $Target. Valid: build, install, test, testacc, fmt, lint, docs, tidy, clean"
        exit 1
    }
}
