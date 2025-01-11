#!/bin/bash

# 检测系统和架构
detect_platform() {
    PLATFORM="unknown"
    ARCH="unknown"

    # 检测操作系统
    case "$(uname -s)" in
        Darwin*)  PLATFORM="darwin";;
        Linux*)   PLATFORM="linux";;
        MINGW64*) PLATFORM="windows";;
        *)        echo "Unsupported platform" && exit 1;;
    esac

    # 检测架构
    case "$(uname -m)" in
        x86_64*)  ARCH="amd64";;
        arm64*)   ARCH="arm64";;
        aarch64*) ARCH="arm64";;
        *)        echo "Unsupported architecture" && exit 1;;
    esac

    if [ "$PLATFORM" = "windows" ]; then
        BINARY_NAME="revid-serve-${PLATFORM}-${ARCH}.exe"
    else
        BINARY_NAME="revid-serve-${PLATFORM}-${ARCH}"
    fi
}

# 下载并安装
install_binary() {
    echo "Downloading revid-serve for $PLATFORM/$ARCH..."
    LATEST_URL="https://github.com/hylarucoder/revid-serve/releases/latest/download/$BINARY_NAME"
    
    # 下载二进制文件
    if ! curl -fsSL "$LATEST_URL" -o "revid-serve"; then
        echo "Download failed"
        exit 1
    fi

    # 添加执行权限
    chmod +x "revid-serve"

    echo "Installation complete! 🎉"
    echo "Run './revid-serve --help' to get started"
}

# 主流程
detect_platform
install_binary
