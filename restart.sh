#!/bin/bash

# 获取 hahub 进程的 PID
pid=$(pgrep -f hahub)

# 如果进程存在,则杀掉它
if [ -n "$pid" ]; then
    echo "Killing hahub process with PID: $pid"
    kill -9 $pid
fi

# 重新以 nohup 后台运行 hahub
echo "Restarting hahub in the background with nohup..."
sudo nohup ./hahub >/dev/null 2>&1 &

echo "hahub process has been restarted in the background."