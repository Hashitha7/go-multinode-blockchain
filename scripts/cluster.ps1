# PowerShell script to start a 3-node blockchain cluster
# Usage: .\scripts\cluster.ps1

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Starting Blockchain Cluster (3 nodes)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Build the node binary
Write-Host "`nBuilding node binary..." -ForegroundColor Yellow
go build -o node.exe ./cmd/node/
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "Build successful!" -ForegroundColor Green

# Start Node 1
Write-Host "`nStarting Node 1 on :3001..." -ForegroundColor Yellow
$node1 = Start-Process -FilePath ".\node.exe" `
    -ArgumentList "--addr", "localhost:3001", "--peers", "localhost:3002,localhost:3003", "--difficulty", "4", "--mine" `
    -PassThru -NoNewWindow -RedirectStandardOutput "node1.log" -RedirectStandardError "node1_err.log"

Start-Sleep -Seconds 1

# Start Node 2
Write-Host "Starting Node 2 on :3002..." -ForegroundColor Yellow
$node2 = Start-Process -FilePath ".\node.exe" `
    -ArgumentList "--addr", "localhost:3002", "--peers", "localhost:3001,localhost:3003", "--difficulty", "4", "--mine" `
    -PassThru -NoNewWindow -RedirectStandardOutput "node2.log" -RedirectStandardError "node2_err.log"

Start-Sleep -Seconds 1

# Start Node 3
Write-Host "Starting Node 3 on :3003..." -ForegroundColor Yellow
$node3 = Start-Process -FilePath ".\node.exe" `
    -ArgumentList "--addr", "localhost:3003", "--peers", "localhost:3001,localhost:3002", "--difficulty", "4", "--mine" `
    -PassThru -NoNewWindow -RedirectStandardOutput "node3.log" -RedirectStandardError "node3_err.log"

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "  Cluster is running!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Nodes:"
Write-Host "  Node 1: http://localhost:3001/status"
Write-Host "  Node 2: http://localhost:3002/status"
Write-Host "  Node 3: http://localhost:3003/status"
Write-Host ""
Write-Host "Useful commands:"
Write-Host "  curl http://localhost:3001/status     # Node status"
Write-Host "  curl http://localhost:3001/chain       # Full chain"
Write-Host "  curl http://localhost:3001/balances    # Account balances"
Write-Host "  curl http://localhost:3001/mempool     # Pending pool"
Write-Host "  curl http://localhost:3001/peers       # Known peers"
Write-Host ""
Write-Host "Press Ctrl+C to stop the cluster..." -ForegroundColor Yellow

# Wait for user to stop
try {
    while ($true) {
        Start-Sleep -Seconds 1
        
        # Check if nodes are still running
        if ($node1.HasExited) { Write-Host "Node 1 has exited!" -ForegroundColor Red }
        if ($node2.HasExited) { Write-Host "Node 2 has exited!" -ForegroundColor Red }
        if ($node3.HasExited) { Write-Host "Node 3 has exited!" -ForegroundColor Red }
    }
}
finally {
    Write-Host "`nStopping cluster..." -ForegroundColor Yellow
    
    if (!$node1.HasExited) { Stop-Process -Id $node1.Id -Force -ErrorAction SilentlyContinue }
    if (!$node2.HasExited) { Stop-Process -Id $node2.Id -Force -ErrorAction SilentlyContinue }
    if (!$node3.HasExited) { Stop-Process -Id $node3.Id -Force -ErrorAction SilentlyContinue }
    
    Write-Host "Cluster stopped." -ForegroundColor Green
}
