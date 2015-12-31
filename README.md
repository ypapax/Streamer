## Overview

A Streamer is a server accepting a stream of UDP packets on one port (`incoming port`) and allowing multiple clients to connect on a different port (`outgoing port`) and receive that stream.

## Incoming stream

The server listens for an incoming stream on a port defined by a command line argument `incoming-port`. any packet received on this port is immediately sent to any connected client.

## Outgoing stream

The server listens for client connections on a port defined by a command line argument `outgoing-port`. clients communicate with the server using the protocol defined below:

### Protocol

`CONNECT id` - connect to server and start receiving the stream

`DISCONNECT id` - disconnect from server

`ALIVE id` - notify server that connection is alive. failing to send this for 30 seconds will cause the server to disconnect that client. usually a client will send this every 10 seconds.

(`id` is a an arbitrary identifier per connection)

