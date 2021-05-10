## General
- Nodes are the core concept of swarm cache. They cache the actual data. 
- Caching is done in form: key=>value. 
- A node may connect to other nodes via websocket, but cannot connect to the same node twice. Multiple nodes interconnected form a "swarm"
- A client may connect to any node via websocket. A client can access the data stored in any nodes swarm to which the connected node is connected to.

A node can :
1. act as client server
2. act as node server (other nodes can connect to it)
3. connect to other node servers

### Client server
- The node starts listening for incoming client connections
- Clients can be any other apps/modules as long as connection is upgraded to websocket and follows the standard is good.
- 

