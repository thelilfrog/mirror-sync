#!/bin/bash

MAKE_PACKAGE=false
TARGET_CURRENT=false

usage() {
 echo "Usage: $0 [OPTIONS]"
 echo "Options:"
 echo " --package             Make a delivery package instead of plain binary"
 echo " --target-current      Target the current OS/arch for the build"
}

# Function to handle options and arguments
handle_options() {
  while [ $# -gt 0 ]; do
    case $1 in
      --package)
        MAKE_PACKAGE=true
        ;;
      --target-current)
        TARGET_CURRENT=true
        ;;
      *)
        echo "Invalid option: $1" >&2
        usage
        exit 1
        ;;
    esac
    shift
  done
}

# Main script execution
handle_options "$@"

if [ ! -d "./build" ]; then
  mkdir ./build
fi

if [ "$TARGET_CURRENT" == "true" ]; then
    GOOS=$(go env GOOS)
    GOARCH=$(go env GOARCH)
    echo "* Compiling daemon for $GOOS/$GOARCH..."

    if [ "$MAKE_PACKAGE" == "true" ]; then
        CGO_ENABLED=0 GORISCV64=rva22u64 GOAMD64=v3 GOARM64=v8.2 go build -o build/mirrorsyncd -a ./cmd/server
        tar -czf build/daemon.tar.gz build/mirrorsyncd
        rm build/mirrorsyncd
    else
      CGO_ENABLED=0 GORISCV64=rva22u64 GOAMD64=v3 GOARM64=v8.2 go build -o build/mirrorsyncd -a ./cmd/server
    fi

    echo "* Compiling client for $GOOS/$GOARCH..."

    if [ "$MAKE_PACKAGE" == "true" ]; then
        CGO_ENABLED=0 GORISCV64=rva22u64 GOAMD64=v3 GOARM64=v8.2 go build -o build/mirrorsync -a ./cmd/cli
        tar -czf build/client.tar.gz build/mirrorsync
        rm build/mirrorsync
    else
      CGO_ENABLED=0  GORISCV64=rva22u64 GOAMD64=v3 GOARM64=v8.2 go build -o build/mirrorsync -a ./cmd/cli
    fi
  exit 0
fi

### FROM HERE, BUILD ALL

## SERVER

platforms=("linux/amd64" "linux/arm64" "linux/riscv64" "linux/ppc64le" "windows/amd64" "darwin/amd64" "darwin/arm64")

for platform in "${platforms[@]}"; do
    echo "* Compiling daemon for $platform..."
    platform_split=(${platform//\// })

    if [ "$MAKE_PACKAGE" == "true" ]; then
        CGO_ENABLED=0 GOOS=${platform_split[0]} GOARCH=${platform_split[1]} GORISCV64=rva22u64 GOAMD64=v3 GOARM64=v8.2 go build -o build/mirrorsyncd -a ./cmd/server
        tar -czf build/daemon_${platform_split[0]}_${platform_split[1]}.tar.gz build/mirrorsyncd
        rm build/mirrorsyncd
    else
      CGO_ENABLED=0 GOOS=${platform_split[0]} GOARCH=${platform_split[1]} GORISCV64=rva22u64 GOAMD64=v3 GOARM64=v8.2 go build -o build/mirrorsyncd_${platform_split[0]}_${platform_split[1]} -a ./cmd/server
    fi
done

for platform in "${platforms[@]}"; do
    echo "* Compiling client for $platform..."
    platform_split=(${platform//\// })

    if [ "$MAKE_PACKAGE" == "true" ]; then
        CGO_ENABLED=0 GOOS=${platform_split[0]} GOARCH=${platform_split[1]} GORISCV64=rva22u64 GOAMD64=v3 GOARM64=v8.2 go build -o build/mirrorsync -a ./cmd/cli
        tar -czf build/client_${platform_split[0]}_${platform_split[1]}.tar.gz build/mirrorsync
        rm build/mirrorsync
    else
      CGO_ENABLED=0 GOOS=${platform_split[0]} GOARCH=${platform_split[1]} GORISCV64=rva22u64 GOAMD64=v3 GOARM64=v8.2 go build -o build/mirrorsync_${platform_split[0]}_${platform_split[1]} -a ./cmd/cli
    fi
done
