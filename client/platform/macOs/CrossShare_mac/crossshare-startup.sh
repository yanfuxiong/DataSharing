#!/bin/bash

# CrossShare Daemon Startup Script
# 以 root 权限在开机时执行

# 日志函数
log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> /var/log/crossshare-daemon.log
}

log_message "CrossShare daemon starting..."

# 等待系统完全启动
sleep 10

# 在这里添加你需要以 root 权限执行的操作
# 例如：设置系统权限、启动后台服务等

# 示例：设置网络权限
# networksetup -setairportnetwork en0 "WiFi名称" "密码"

# 示例：启动特定服务
# launchctl load -w /System/Library/LaunchDaemons/some.service.plist

# 示例：设置文件权限
# chown -R root:wheel /usr/local/bin/crossshare/
# chmod 755 /usr/local/bin/crossshare/*

# 你可以在这里添加 CrossShare 需要的特定配置
log_message "Configuring CrossShare system settings..."

# 示例：配置防火墙规则
# pfctl -f /etc/pf.conf

# 示例：设置系统参数
# sysctl -w kern.maxfiles=65536

# 检查 CrossShare 应用是否存在
CROSSSHARE_APP="/Applications/CrossShare.app"
if [ -d "$CROSSSHARE_APP" ]; then
    log_message "CrossShare app found at $CROSSSHARE_APP"
    
    # 可以在这里启动应用或配置相关权限
    # 注意：GUI 应用通常不应在这里启动，而是通过用户会话启动
    
    # 配置应用权限
    chown -R root:admin "$CROSSSHARE_APP"
    chmod -R 755 "$CROSSSHARE_APP"
    
else
    log_message "Warning: CrossShare app not found at $CROSSSHARE_APP"
fi

log_message "CrossShare daemon initialization completed"

exit 0