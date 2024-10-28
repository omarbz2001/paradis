# Depth-ST-Without-Root
This project implements a distributed algorithm in Golang to determine a leader node in a network, select parents, and propagate a termination signal using a simple message-passing model. Nodes communicate over TCP, sending messages to neighbors based on an initial configuration.



## Table of Contents

- Project Overview
- Requirements
- Setup
- Configuration
- Running the Program
- Message Types
- Logging



## Project Overview

This project is a simulated network of nodes that:

1- Elects a leader node based on ID values.
2- Selects parent nodes for each node.
3- Terminates with a broadcast message once the leader node is identified.

Each node starts a TCP server, allowing nodes to communicate with their neighbors as defined in a YAML configuration file.

## Requirements

Golang (tested on Go version 1.17+)
YAML parsing library (gopkg.in/yaml.v2)

## Setup

1- Clone the repository:

```bash
git clone https://github.com/omarbz2001/paradis
cd depth-ST-without-root
```
2- Install dependencies: This project uses YAML for configuration parsing. Run:
```bash
go get gopkg.in/yaml.v2
```

***

# Configuration
Each node's properties are defined in separate YAML configuration files (e.g., node-1.yaml, node-2.yaml, etc.) .

Each YAML file has the following structure:
```bash
id: <Node ID>
address: <IP Address>
neighbours:
  - id: <Neighbor Node ID>
    address: <Neighbor IP Address>
    edge_weight: <Weight of Edge>
```
Example : 
```bash
id: 1
address: "localhost:30001"
neighbours:
  - id: 2
    address: "localhost:30002"
    edge_weight: 1
  - id: 3
    address: "localhost:30003"
    edge_weight: 1
```

## Running the Program
The program initializes each node server using a separate goroutine. To start all nodes:
```bash
go run construct-ST.go
```
Each node's log messages will be saved in separate files with a filename pattern Log-<Node_Address>. 
Each node writes its interactions to this log, detailing sent and received messages and status updates.

## Message Types
The program uses the following message types for communication:

M (Message): Used by nodes to propose a parent. A node with the highest ID becomes the leader.
P (Parent): Confirmation of parent selection.
R (Reject): Response for nodes that do not become parents.
T (Terminate): Sent by the root node to indicate the end of the algorithm.
## Logging
Each node creates a unique log file named Log-<Node_Address> containing all the communication steps, parent selection, leader election, and termination status.
## Example Log
```bash
Parsing done ...
Server starting ....
Starting algorithm ...
Sent initial <M, 1> to node 2
Received <M, 1> from node 2
Adopting node 2 as parent
Sent <P, 1> to node 2
```



