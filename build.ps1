$BINARY_NAME = "devflow"
$VERSION = "0.1.0"

$gitCommit = git rev-parse --short HEAD 2>$null
if ($LASTEXITCODE -ne 0) { $gitCommit = "none" }

$DATE = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$LDFLAGS = "-X main.Version=$VERSION -X main.Commit=$gitCommit -X main.Date=$DATE"

function Build {
    if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
    go build -ldflags $LDFLAGS -o "bin/$BINARY_NAME.exe" ./cmd/devflow
    if ($LASTEXITCODE -eq 0) { Write-Host "Build succeeded: bin/$BINARY_NAME.exe" -ForegroundColor Green }
}

function Test {
    go test -v ./internal/...
}

function TestCoverage {
    go test -coverprofile=coverage.out ./internal/...
    go tool cover -html=coverage.out
}

function Clean {
    if (Test-Path "bin") { Remove-Item -Recurse -Force "bin" }
    if (Test-Path "coverage.out") { Remove-Item -Force "coverage.out" }
}

function Install {
    go install -ldflags $LDFLAGS ./cmd/devflow
}

function Run {
    go run ./cmd/devflow @args
}

function BuildAll {
    if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }

    $env:GOOS = "linux";   $env:GOARCH = "amd64"; go build -ldflags $LDFLAGS -o "bin/$BINARY_NAME-linux-amd64"     ./cmd/devflow
    $env:GOOS = "linux";   $env:GOARCH = "arm64"; go build -ldflags $LDFLAGS -o "bin/$BINARY_NAME-linux-arm64"     ./cmd/devflow
    $env:GOOS = "darwin";  $env:GOARCH = "amd64"; go build -ldflags $LDFLAGS -o "bin/$BINARY_NAME-darwin-amd64"    ./cmd/devflow
    $env:GOOS = "darwin";  $env:GOARCH = "arm64"; go build -ldflags $LDFLAGS -o "bin/$BINARY_NAME-darwin-arm64"    ./cmd/devflow
    $env:GOOS = "windows"; $env:GOARCH = "amd64"; go build -ldflags $LDFLAGS -o "bin/$BINARY_NAME-windows-amd64.exe" ./cmd/devflow

    $env:GOOS = ""; $env:GOARCH = ""
    Write-Host "Build all platforms succeeded" -ForegroundColor Green
}

switch ($args[0]) {
    "build"          { Build }
    "test"           { Test }
    "test-coverage"  { TestCoverage }
    "clean"          { Clean }
    "install"        { Install }
    "run"            { Run $args[1..($args.Length-1)] }
    "build-all"      { BuildAll }
    default          { Write-Host "Usage: .\build.ps1 <build|test|test-coverage|clean|install|run|build-all>" }
}
