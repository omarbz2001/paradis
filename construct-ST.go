package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var PORT string = ":30000"

type Neighbour struct {
	ID         int    `yaml:"id"`
	Address    string `yaml:"address"`
	EdgeWeight int    `yaml:"edge_weight"`
}

type yamlConfig struct {
	ID         int         `yaml:"id"`
	Address    string      `yaml:"address"`
	Neighbours []Neighbour `yaml:"neighbours"`
}

func initAndParseFileNeighbours(filename string) yamlConfig {
	fullpath, _ := filepath.Abs("./" + filename)
	yamlFile, err := ioutil.ReadFile(fullpath)

	if err != nil {
		panic(err)
	}

	var data yamlConfig

	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		panic(err)
	}

	return data
}

func Log(file *os.File, message string) {
	_, err := file.WriteString(message)
	if err != nil {
		panic(err)
	}
}

func send(nodeConfig yamlConfig, neighAddress string, msg string) {
    remoteAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s%s", neighAddress, PORT))
    if err != nil {
        log.Printf("Failed to resolve remote address: %s", err)
        return
    }

    outConn, err := net.DialTCP("tcp", nil, remoteAddr)
    if err != nil {
        log.Printf("Failed to connect to %s from node %d: %s", remoteAddr, nodeConfig.ID, err)
        return
    }
    defer outConn.Close()

    outConn.Write([]byte(msg + "\n"))
    log.Printf("Node %d sent message to %s", nodeConfig.ID, remoteAddr.String())
}


func server(neighboursFilePath string) {
	nodeConfig := initAndParseFileNeighbours(neighboursFilePath)
	var node yamlConfig = initAndParseFileNeighbours(neighboursFilePath)
	Leader := nodeConfig.ID
	Parent := -1
	F := []int{}
	NF := []int{}
	NE := nodeConfig.Neighbours

	filename := "Log-" + nodeConfig.Address
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	Log(file, "Parsing done ...\n")
	Log(file, "Server starting ....\n")

	ln, err := net.Listen("tcp", node.Address+PORT)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ln.Close()

	var terminated = false
	Log(file, "Starting algorithm ...\n")

	time.Sleep(2 * time.Second)

	if len(NE) > 0 {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(len(NE))
		neigh := NE[randIndex]
		send(nodeConfig, neigh.Address, "M\t"+strconv.Itoa(nodeConfig.ID))
		Log(file, "Sent initial <M, "+strconv.Itoa(nodeConfig.ID)+"> to node "+strconv.Itoa(neigh.ID)+"\n")
		NE = append(NE[:randIndex], NE[randIndex+1:]...)
	}
	fmt.Println(terminated)

	for !terminated {
		conn, _ := ln.Accept()
		message, _ := bufio.NewReader(conn).ReadString('\n')
		conn.Close()

		Log(file, "Message received: "+message+"\n")
		message = strings.TrimSpace(message)
		parts := strings.Split(message, "\t")

		msgType := parts[0]
		senderID, _ := strconv.Atoi(parts[1])

		if msgType == "M" {
			Log(file, "Received <M, "+strconv.Itoa(senderID)+"> from node "+strconv.Itoa(senderID)+"\n")

			if Parent == -1 {
				if senderID > Leader {
					Leader = senderID
					Parent = senderID
					Log(file, "Adopting node "+strconv.Itoa(senderID)+" as parent\n")
					F = nil
					NF = nil
				}
			}

			if len(NE) > 0 {
				neigh := NE[0]
				send(nodeConfig, neigh.Address, "M\t"+strconv.Itoa(Leader))
				NE = NE[1:]
			} else {
				if Parent != -1 && Parent != nodeConfig.ID && len(NE) > 0 {
					send(nodeConfig, NE[0].Address, "P\t"+strconv.Itoa(Leader))
				}
			}
			
		}

		if msgType == "P" || msgType == "R" {
			Log(file, "Received <"+msgType+", "+strconv.Itoa(senderID)+">\n")

			if senderID == Leader {
				if msgType == "R" {
					NF = append(NF, senderID)
				} else if msgType == "P" {
					F = append(F, senderID)
				}

				if len(NE) > 0 {
					neigh := NE[0]
					send(nodeConfig, neigh.Address, "M\t"+strconv.Itoa(Leader))
					NE = NE[1:]
				} else {
					if Parent != -1 && Parent != nodeConfig.ID {
						send(nodeConfig, NE[0].Address, "P\t"+strconv.Itoa(Leader))
					} else {
						Log(file, "I am the root node.\n")
						for _, neighbor := range NE {
							send(nodeConfig, neighbor.Address, "T\t"+strconv.Itoa(Leader))
						}
						terminated = true
					}
				}
			}
		}

		if msgType == "T" {
			Log(file, "Received <T, "+strconv.Itoa(senderID)+"> - Termination initiated by root\n")
			for _, neighbor := range NE {
				send(nodeConfig, neighbor.Address, "T\t"+strconv.Itoa(Leader))
			}
			terminated = true
		}
	}
}

func main() {
	go server("node-2.yaml")
	go server("node-3.yaml")
	go server("node-4.yaml")
	go server("node-5.yaml")
	go server("node-6.yaml")
	go server("node-7.yaml")
	go server("node-8.yaml")
	time.Sleep(2 * time.Second)
	server("node-1.yaml")
	time.Sleep(2 * time.Second)
}
