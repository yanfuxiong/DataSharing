#!/bin/bash

# CrossShare Launch Daemon install script
# 在 Xcode Build Phase 中作为 Run Script 执行

set -e

# 检查运行模式
DRY_RUN=false
INSTALL_MODE=false

if [ "$1" == "--dry-run" ]; then
    DRY_RUN=true
    echo "🔍 Validating CrossShare Launch Daemon (Dry Run Mode)..."
elif [ "$1" == "--install" ]; then
    INSTALL_MODE=true
    echo "📦 Installing CrossShare Launch Daemon..."
else
    echo "Usage: $0 [--dry-run|--install]"
    exit 1
fi

SOURCE_DIR="$SRCROOT"
PLIST_SOURCE="$SOURCE_DIR/com.instance.crossshare.identifier.macos.daemon.plist"
SCRIPT_SOURCE="$SOURCE_DIR/crossshare-startup.sh"

# 目标路径
if [ "$DRY_RUN" = true ]; then
    # Dry run 模式：使用临时目录
    TEMP_DIR="/tmp/crossshare-daemon-test"
    PLIST_DEST="$TEMP_DIR/LaunchDaemons/com.instance.crossshare.identifier.macos.daemon.plist"
    SCRIPT_DEST="$TEMP_DIR/bin/crossshare-startup.sh"
else
    # 安装模式：使用系统目录
    PLIST_DEST="/Library/LaunchDaemons/com.instance.crossshare.identifier.macos.daemon.plist"
    SCRIPT_DEST="/usr/local/bin/crossshare-startup.sh"
fi

echo "🔍 Checking source files..."

if [ ! -f "$PLIST_SOURCE" ]; then
    echo "❌ Error: plist file not found at $PLIST_SOURCE"
    exit 1
fi
echo "✅ Found plist file: $PLIST_SOURCE"

if [ ! -f "$SCRIPT_SOURCE" ]; then
    echo "❌ Error: shell script not found at $SCRIPT_SOURCE"
    exit 1
fi
echo "✅ Found shell script: $SCRIPT_SOURCE"

# 验证 plist 文件语法
echo "🔍 Validating plist syntax..."
if plutil -lint "$PLIST_SOURCE" > /dev/null; then
    echo "✅ plist syntax is valid"
else
    echo "❌ plist syntax error"
    exit 1
fi

# 验证 shell 脚本语法
echo "🔍 Validating shell script syntax..."
if bash -n "$SCRIPT_SOURCE"; then
    echo "✅ Shell script syntax is valid"
else
    echo "❌ Shell script syntax error"
    exit 1
fi

# 创建目录
if [ "$DRY_RUN" = true ]; then
    echo "🔍 Creating temporary directories..."
    mkdir -p "$(dirname "$PLIST_DEST")"
    mkdir -p "$(dirname "$SCRIPT_DEST")"
    mkdir -p "$TEMP_DIR/var/log"
else
    echo "📁 Creating system directories..."
    sudo mkdir -p /usr/local/bin
    sudo mkdir -p /var/log
fi

echo "📋 Processing files..."

# 复制 shell 脚本
if [ "$DRY_RUN" = true ]; then
    cp "$SCRIPT_SOURCE" "$SCRIPT_DEST"
    chmod 755 "$SCRIPT_DEST"
    echo "✅ (Dry Run) Shell script copied to $SCRIPT_DEST"
else
    sudo cp "$SCRIPT_SOURCE" "$SCRIPT_DEST"
    sudo chown root:wheel "$SCRIPT_DEST"
    sudo chmod 755 "$SCRIPT_DEST"
    echo "✅ Shell script installed to $SCRIPT_DEST"
fi

# 复制 plist 文件
if [ "$DRY_RUN" = true ]; then
    cp "$PLIST_SOURCE" "$PLIST_DEST"
    chmod 644 "$PLIST_DEST"
    echo "✅ (Dry Run) plist file copied to $PLIST_DEST"
else
    sudo cp "$PLIST_SOURCE" "$PLIST_DEST"
    sudo chown root:wheel "$PLIST_DEST"
    sudo chmod 644 "$PLIST_DEST"
    echo "✅ plist file installed to $PLIST_DEST"
fi

# 创建日志文件
if [ "$DRY_RUN" = true ]; then
    touch "$TEMP_DIR/var/log/crossshare-daemon.log"
    touch "$TEMP_DIR/var/log/crossshare-daemon-error.log"
    chmod 644 "$TEMP_DIR/var/log/crossshare-daemon"*.log
    echo "✅ (Dry Run) Log files created in $TEMP_DIR/var/log/"
else
    sudo touch /var/log/crossshare-daemon.log
    sudo touch /var/log/crossshare-daemon-error.log
    sudo chown root:wheel /var/log/crossshare-daemon*.log
    sudo chmod 644 /var/log/crossshare-daemon*.log
    echo "✅ Log files created"
fi

# 处理 launchctl 操作
if [ "$DRY_RUN" = true ]; then
    echo "🔍 (Dry Run) Would check for existing daemon..."
    echo "🔍 (Dry Run) Would load daemon from: $PLIST_DEST"
    echo "✅ (Dry Run) All files validated and would be installed successfully"
else
    if sudo launchctl list | grep -q "com.instance.crossshare.identifier.macos.daemon"; then
        echo "🔄 Unloading existing daemon..."
        sudo launchctl unload "$PLIST_DEST" 2>/dev/null || true
    fi

    echo "🚀 Loading daemon..."
    sudo launchctl load "$PLIST_DEST"

    if sudo launchctl list | grep -q "com.instance.crossshare.identifier.macos.daemon"; then
        echo "✅ CrossShare daemon loaded successfully"
    else
        echo "⚠️ Warning: Daemon may not have loaded properly"
    fi
fi

if [ "$DRY_RUN" = true ]; then
    echo "🎉 CrossShare Launch Daemon validation completed!"
    echo ""
    echo "📊 Validation Summary:"
    echo "✅ Source files exist and are valid"
    echo "✅ plist syntax is correct"
    echo "✅ Shell script syntax is correct"
    echo "✅ All files would be installed to correct locations"
    echo ""
    echo "🗂️ Test Files Created:"
    echo "  - plist: $PLIST_DEST"
    echo "  - script: $SCRIPT_DEST"
    echo "  - logs: $TEMP_DIR/var/log/crossshare-daemon*.log"
    echo ""
    echo "🧹 Cleaning up test files..."
    rm -rf "$TEMP_DIR"
    echo "✅ Test files cleaned up"
else
    echo "🎉 CrossShare Launch Daemon installation completed!"
    echo ""
    echo "📊 Daemon Status:"
    sudo launchctl list | grep crossshare || echo "No crossshare daemons found"
fi

# 显示管理命令（仅在实际安装时）
if [ "$INSTALL_MODE" = true ]; then
    echo ""
    echo "🔧 管理命令:"
    echo "  加载: sudo launchctl load $PLIST_DEST"
    echo "  卸载: sudo launchctl unload $PLIST_DEST"
    echo "  查看日志: tail -f /var/log/crossshare-daemon.log"
fi
