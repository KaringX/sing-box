---
icon: material/new-box
---

### 结构

```json
{
  "type": "mieru",
  "tag": "mieru-out",

  "server": "127.0.0.1",
  "server_port": 1080,
  "server_ports": [
    "9000-9010",
    "9020-9030"
  ],
  "transport": "TCP",
  "username": "asdf",
  "password": "hjkl",
  "multiplexing": "MULTIPLEXING_LOW",
  "traffic_pattern": "GgQIARAK",
  "mtu": 1400,
  "handshake_mode": "HANDSHAKE_STANDARD",

  ... // 拨号字段
}
```

### 字段

#### server

==必填==

服务器地址。

#### server_port

服务器端口。

必须填写 `server_port` 和 `server_ports` 中至少一项。

#### server_ports

服务器端口范围列表。

必须填写 `server_port` 和 `server_ports` 中至少一项。

#### transport

==必填==

通信协议。仅可设为 `TCP`。

#### username

==必填==

mieru 用户名。

#### password

==必填==

mieru 密码。

#### multiplexing

多路复用设置。可以设为 `MULTIPLEXING_OFF`，`MULTIPLEXING_LOW`，`MULTIPLEXING_MIDDLE`，`MULTIPLEXING_HIGH`。其中 `MULTIPLEXING_OFF` 会关闭多路复用功能。

#### traffic_pattern

一个 base64 字符串用于微调网络行为。

#### mtu

L2 载荷的最大传输单元。仅适用于 UDP 传输方式的出站流量。

#### handshake_mode

客户端建立连接时的握手模式。可设为 `HANDSHAKE_STANDARD`（1-RTT，客户端等待代理服务器建立到目标地址的连接后再发送载荷）或 `HANDSHAKE_NO_WAIT`（0-RTT，客户端在连接代理服务器的同时发送载荷）。

### 拨号字段

参阅 [拨号字段](/zh/configuration/shared/dial/)。
