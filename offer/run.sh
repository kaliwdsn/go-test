#!/bin/bash

# UDP监听地址
listen_addr="127.0.0.1:8989"

# 创建UDP监听
exec 3<>/dev/udp/$listen_addr

# 读取并打印接收到的结果
while read -r -u 3 result; do
  echo "Received result: $result"
done
