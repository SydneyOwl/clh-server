# clh-server [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/SydneyOwl/clh-server)

`clh-server` is the server side of Cloudlog Helper. 
It caches or forwards messages from WSJT-X or similar applications to multiple compatible clients that subscribe to specific endpoints.

## Overview

Riginfo/WSJT-X Info -> clh-client (sender) -> clh-server -> clh-client (receiver) -> multiple endpoints

`Riginfo/WSJT-X Info` can be provided by cloudlog-helper or your own application as long as you impl the packet format below.

Or you can use the example sender program (`./clh-server run-receiver --tls --key=xxxx --skip-cert-verify --ip=xxx --port=xxx`).
This provides an udp server which recv udp packets from wsjt-x or jtdx.

## Packet format

Looks exactly like frp control packets:

```
+--------+--------+--------+--------+--------+------------------+  
| Type     | Length (32-bit big endian)      | Payload          |  
| (1 byte) |                                 | (protobuf bytes) |  
+--------+--------+--------+--------+--------+------------------+


Type Byte Mapping:  
'1' = Ping  
'2' = Pong    
'3' = Command  
'4' = CommonResponse  
'5' = CommandResponse  
'a' = HandshakeRequest  
'b' = HandshakeResponse  
'c' = WsjtxMessage  
'd' = WsjtxMessagePacked
```

## Protocols / workflows
> [!important]
> See proto files for more.

### Handshake Flow
Client sends HandshakeRequest with:
+ ClientType: "sender" or "receiver"
+ AuthKey: Calculated using shared secret and timestamp
+ Timestamp: Current Unix time
+ RunId: Unique client identifier

The Server validates authentication and responds with HandshakeResponse containing the accepted RunId.

## Installation

1. Clone the repository recursively (--recursive)
2. Run `make init` to install dependencies
3. Run `make proto` to generate protobuf files
4. Run `make build` to build the binary

## Usage

1. Initialize config: `clh-server init-config`
2. Check config: `clh-server check-config`
3. Run server: `clh-server`

## Configuration

See the generated config.yaml for options.

## Commands

- `init-config`: Generate default config
- `check-config`: Validate config
- `run-receiver`: Run example receiver client
- `run-sender`: Run example sender client

## Important Notices

+ TLS is enabled by default but can be disabled through configuration. 
Disabling TLS is strongly discouraged as it exposes all communication including authentication keys in plain text. 
+ If certificate files are not provided, the server will automatically generate self-signed certificates. 
+ TLS certificate verification / client-side cert is not supported so far. This means **the connection vulnerable to man-in-the-middle attacks**. 