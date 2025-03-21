# p2p chat app with libp2p [with peer discovery]
(Source code adapted from a libp2p [examples](https://github.com/libp2p/go-libp2p/tree/master/examples/chat-with-rendezvous))

This program demonstrates a simple p2p chat application. 
You will learn how to discover a peer in the network (using kad-dht), connect to it and open a chat stream.

To recap the current example:
1. uses ipfs bootstrap nodes to connect to the p2p network
2. identifies itself using an Ed25519 key
3. looks for other nodes that can handle a brand new "/p2p-rag/0.0.0" protocol
4. "chats" with the discovered nodes.

You need go 1.24 for building this program. You can install it directly on your machine, or use it via docker:
```bash
docker run --rm -it -v $PWD:/app -w /app golang:1.24.1 bash
```

## Build

From the `go` directory (i.e. the directory containing this file) run the following:

```
> go build
```

## Usage

Use two different terminal windows to run

```
./p2p-rag -listen /ip4/127.0.0.1/tcp/6666 -rendezvous someString
./p2p-rag -listen /ip4/127.0.0.1/tcp/6668 -rendezvous someString
```

It would be better if each session uses its own private key each time:
```
./p2p-rag -listen /ip4/127.0.0.1/tcp/6666 -rendezvous someString -key CAESQBEx8bKxlGKCzcfxsR94EEqCE+8bBm/fBaAXOBEkXiU9uGjKXsoDUJJuGugLJFVvbwvbqeZzP0IR23M45C0mkRY=
./p2p-rag -listen /ip4/127.0.0.1/tcp/6668 -rendezvous someString -key CAESQHKQORG8iqd0O3fHj16d4KPA4Y2NhaVzEafylicBx92b2UqHPw4Zjss+BymIX4oqkSzfqQHYwVVtE6fNSLwstlY=
```

To obtain a new private key run:
```
./p2p-rag -pk
```

## So how does it work?

1. **Configure a p2p host**
```go
// libp2p.New constructs a new libp2p Host.
// Other options can be added here.
host, err := libp2p.New()
```
[libp2p.New](https://pkg.go.dev/github.com/libp2p/go-libp2p#New) is the constructor for a libp2p node. It creates a host with the given configuration. Right now, all the options are default, documented [here](https://pkg.go.dev/github.com/libp2p/go-libp2p#New)

2. **Set a default handler function for incoming connections.**

This function is called on the local peer when a remote peer initiates a connection and starts a stream with the local peer.
```go
// Set a function as stream handler.
host.SetStreamHandler("/p2p-rag/0.0.0", handleStream)
```

```handleStream``` is executed for each new incoming stream to the local peer. ```stream``` is used to exchange data between the local and remote peers. This example uses non blocking functions for reading and writing from this stream.

```go
func handleStream(stream net.Stream) {

    // Create a buffer stream for non blocking read and write.
    rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

    go readData(rw)
    go writeData(rw)

    // 'stream' will stay open until you close it (or the other side closes it).
}
```

3. **Initiate a new DHT Client with ```host``` as local peer.**


```go
dht, err := dht.New(ctx, host)
```

4. **Connect to IPFS bootstrap nodes.**

These nodes are used to find nearby peers using DHT.

```go
for _, addr := range bootstrapPeers {

    iaddr, _ := ipfsaddr.ParseString(addr)

    peerinfo, _ := peerstore.InfoFromP2pAddr(iaddr.Multiaddr())

    if err := host.Connect(ctx, *peerinfo); err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("Connection established with bootstrap node: ", *peerinfo)
    }
}
```

5. **Announce your presence using a rendezvous point.**

[routingDiscovery.Advertise](https://pkg.go.dev/github.com/libp2p/go-libp2p/p2p/discovery/routing#RoutingDiscovery.Advertise) makes this node announce that it can provide a value for the given key. Where a key in this case is ```rendezvousString```. Other peers will hit the same key to find other peers.

```go
routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
discovery.Advertise(ctx, routingDiscovery, config.RendezvousString)
```

6. **Find nearby peers.**

[routingDiscovery.FindPeers](https://pkg.go.dev/github.com/libp2p/go-libp2p/p2p/discovery/routing#RoutingDiscovery.FindPeers) will return a channel of peers who have announced their presence.

```go
peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
```

The [discovery](https://pkg.go.dev/github.com/libp2p/go-libp2p/p2p/discovery/routing) package uses the DHT internally to [provide](https://pkg.go.dev/github.com/libp2p/go-libp2p-kad-dht#IpfsDHT.Provide) and [findProviders](https://pkg.go.dev/github.com/libp2p/go-libp2p-kad-dht#IpfsDHT.FindProviders).

**Note:** Although [routingDiscovery.Advertise](https://pkg.go.dev/github.com/libp2p/go-libp2p/p2p/discovery/routing#RoutingDiscovery.Advertise) and [routingDiscovery.FindPeers](https://pkg.go.dev/github.com/libp2p/go-libp2p/p2p/discovery/routing#RoutingDiscovery.FindPeers) works for a rendezvous peer discovery, this is not the right way of doing it. Libp2p is currently working on an actual rendezvous protocol ([libp2p/specs#56](https://github.com/libp2p/specs/pull/56)) which can be used for bootstrap purposes, real time peer discovery and application specific routing.

7. **Open streams to newly discovered peers.**

Finally we open streams to the newly discovered peers.

```go
go func() {
		for peer := range peerChan {
			if peer.ID == host.ID() {
				continue
			}
			fmt.Println("Found peer:", peer)

			fmt.Println("Connecting to:", peer)
			stream, err := host.NewStream(ctx, peer.ID, protocol.ID(config.ProtocolID))

			if err != nil {
				fmt.Println("Connection failed:", err)
				continue
			} else {
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

				go writeData(rw)
				go readData(rw)
			}

			fmt.Println("Connected to:", peer)
		}
	}()
```

## Authors
1. Abhishek Upperwal
2. Mantas Vidutis
