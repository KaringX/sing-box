---
icon: material/new-box
---

### Structure

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

  ... // Dial Fields
}
```

### Fields

#### server

==Required==

The server address.

#### server_port

The server port.

Must set at least one field between `server_port` and `server_ports`.

#### server_ports

Server port range list.

Must set at least one field between `server_port` and `server_ports`.

#### transport

==Required==

Transmission protocol. The only allowed value is `TCP`.

#### username

==Required==

mieru user name.

#### password

==Required==

mieru password.

#### multiplexing

Multiplexing level. Supported values are `MULTIPLEXING_OFF`, `MULTIPLEXING_LOW`, `MULTIPLEXING_MIDDLE`, `MULTIPLEXING_HIGH`. `MULTIPLEXING_OFF` disables multiplexing.

#### traffic_pattern

A base64 string to fine tune network behavior.

#### mtu

Maximum transmission unit of L2 payload. Only applies to UDP transport egress traffic.

#### handshake_mode

Handshake mode when the client opens a connection. Allowed values are `HANDSHAKE_STANDARD` (1-RTT, wait for the proxy server to establish the connection to the destination before sending payload) and `HANDSHAKE_NO_WAIT` (0-RTT, send payload together with the connection to the proxy server).

### Dial Fields

See [Dial Fields](/configuration/shared/dial/) for details.
