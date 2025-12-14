# go-zenon FTW (For The Windows)

A community Windows port of [go-zenon](https://github.com/zenon-network/go-zenon) - the reference Go implementation of the Zenon Network of Momentum.

**No C compiler required** - this port uses pure Go dependencies for easy Windows builds.

## Quick Start

1. Download `znnd.exe` from the [releases page](https://github.com/wave-func/go-zennon-ftw/releases/tag/v0.0.8-windows)
2. Run it:
```
znnd.exe
```

That's it! The node will sync with the Alphanet network.

## Building from Source

```
go build -o build/znnd.exe ./cmd/znnd
```

## Windows Port Changes

### 1. Pure Go Crypto Library
Replaced CGO-dependent `go-ethereum/crypto/secp256k1` with pure Go `decred/dcrd/dcrec/secp256k1/v4`.

| Files Modified |
|----------------|
| `p2p/rlpx.go`, `p2p/discover/node.go` |
| `vm/embedded/implementation/bridge.go`, `vm/embedded/implementation/swap_utils.go` |

### 2. Platform-Specific Signal Handling
Windows doesn't support `SIGKILL`. Added build-tagged signal files.

| File | Platform |
|------|----------|
| `app/signals_unix.go` | Unix (SIGINT, SIGTERM, SIGKILL) |
| `app/signals_windows.go` | Windows (SIGINT, SIGTERM) |

### 3. Platform-Specific Error Codes
Unix errno values differ from Windows error codes.

| File | Platform |
|------|----------|
| `node/errors_unix.go` | Unix errnos (11, 32, 35) |
| `node/errors_windows.go` | Windows error codes (32, 33) |

### 4. Updated gopsutil
Upgraded from `v3.21.11` to `v3.24.5` for improved Windows support.

### 5. Fixed Path Separators
Replaced `path.Join` with `filepath.Join` for OS-native path handling.

### 6. Fixed Home Directory Detection
Windows doesn't set `HOME` env var. Now uses `user.Current()` first.

### 7. Replaced Deprecated API
Updated `Readdir` to `ReadDir` in `node/defaults.go`.

## Sync Optimizations

Optional performance tuning included:

| Parameter | Original | New |
|-----------|----------|-----|
| `maxBlockProcess` | 256 | 512 |
| Ticker interval | 100ms | 50ms |
| `blockSoftTTL` | 3s | 4s |
| `blockCacheLimit` | 32 * 128 | 64 * 128 |
| Starting peer capacity | 1 | 16 |
| `forceSyncCycle` | 4s | 2s |

## Credits

Based on [go-zenon](https://github.com/zenon-network/go-zenon) by Zenon Network.
