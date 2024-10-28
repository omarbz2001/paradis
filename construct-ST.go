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

func sendToAllNeighbors(nodeConfig yamlConfig, neighbors []Neighbour, messageType string) {
	for _, neigh := range neighbors {
		msg := "T\t" + strconv.Itoa(nodeConfig.ID) + " from " + strconv.Itoa(nodeConfig.ID) + "\n"
		send(nodeConfig, neigh.Address, msg)
		log.Printf("Sent <%s, %d> to node %d\n", messageType, nodeConfig.ID, neigh.ID)
	}
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
	var i int = 0
	Log(file, "Starting algorithm ...\n")

	time.Sleep(2 * time.Second)

	if len(NE) > 0 {
		rand.Seed(time.Now().UnixNano())
		randIndex := rand.Intn(len(NE))
		neigh := NE[randIndex]
		send(nodeConfig, neigh.Address, "M\t"+strconv.Itoa(nodeConfig.ID)+" from "+strconv.Itoa(nodeConfig.ID)+"\n")
		Log(file, "Sent initial <M, "+strconv.Itoa(nodeConfig.ID)+"> to node "+strconv.Itoa(neigh.ID)+"\n")
		NE = append(NE[:randIndex], NE[randIndex+1:]...)
	}

	for !terminated {
		conn, _ := ln.Accept()
		message, _ := bufio.NewReader(conn).ReadString('\n')
		conn.Close()

		fmt.Println(message)
		message = strings.TrimSpace(message)
		parts := strings.Fields(message)

		msgType := parts[0]
		senderID, _ := strconv.Atoi(parts[1])
		neighID, _ := strconv.Atoi(parts[3])

		if msgType == "M" {
			Log(file, "Received <"+msgType+", "+strconv.Itoa(senderID)+"> from node "+strconv.Itoa(neighID)+"\n")

			if senderID > Leader {
				Leader = senderID
				Parent = senderID
				Log(file, "Adopting node "+strconv.Itoa(neighID)+" as parent\n")

				for _, neigh := range NE {
					if neigh.ID == Parent {
						send(nodeConfig, neigh.Address, "P\t"+strconv.Itoa(Leader)+" from "+strconv.Itoa(nodeConfig.ID))
						Log(file, "Sent <P, "+strconv.Itoa(nodeConfig.ID)+"> to node "+strconv.Itoa(neigh.ID)+"\n")
						break
					}
				}
				F = nil
				NF = nil
			} else {
				for _, neigh := range NE {
					if neigh.ID == Parent {
						send(nodeConfig, neigh.Address, "R\t"+strconv.Itoa(Leader)+" from "+strconv.Itoa(nodeConfig.ID))
						Log(file, "Sent <R, "+strconv.Itoa(nodeConfig.ID)+"> to node "+strconv.Itoa(neigh.ID)+"\n")
						break
					}
				}
			}
			if len(NE) > 0 {
				neigh := NE[0]
				send(nodeConfig, neigh.Address, "M\t"+strconv.Itoa(Leader)+" from "+strconv.Itoa(nodeConfig.ID))
				NE = NE[1:]
			}

		}

		if msgType == "P" || msgType == "R" {
			Log(file, "Received <"+msgType+", "+strconv.Itoa(senderID)+">\t"+"> from node "+strconv.Itoa(senderID)+"\n")

			if senderID == Leader {
				if msgType == "R" {
					NF = append(NF, senderID)
					i = i + 1
					for i, neigh := range NE {
						if neigh.ID == senderID {
							NE = append(NE[:i], NE[i+1:]...)
							break
						}
					}
				} else if msgType == "P" {
					F = append(F, senderID)
					i = i + 1
					for i, neigh := range NE {
						if neigh.ID == senderID {
							NE = append(NE[:i], NE[i+1:]...)
						}
					}
				}
			}
			if senderID > Leader {
				i = 0
				if msgType == "R" {
					NF = append(NF, senderID)
					i = i + 1
					for i, neigh := range NE {
						if neigh.ID == senderID {
							NE = append(NE[:i], NE[i+1:]...)
							break
						}
					}
				} else if msgType == "P" {
					F = append(F, senderID)
					i = i + 1
					for i, neigh := range NE {
						if neigh.ID == senderID {
							NE = append(NE[:i], NE[i+1:]...)
						}
					}
				}
			}

			if len(NE) > 0 {
				neigh := NE[0]
				send(nodeConfig, neigh.Address, "M\t"+strconv.Itoa(Leader)+" from "+strconv.Itoa(nodeConfig.ID))
				NE = NE[1:]
			} else {
				numNeighbors := len(NF) + len(F)
				if i == numNeighbors && nodeConfig.ID > 7 {
					sendToAllNeighbors(nodeConfig, nodeConfig.Neighbours, "T")
					Log(file, "I am node root"+"\n")
					Log(file, "Sent T to all neighbors"+"\n")
					terminated = true
				}
			}

		}

		if msgType == "T" {
			Log(file, "Received <T, "+strconv.Itoa(senderID)+"> - Termination initiated by root\n")
			for _, neighbor := range nodeConfig.Neighbours {
				send(nodeConfig, neighbor.Address, "T\t"+strconv.Itoa(Leader)+" from "+strconv.Itoa(nodeConfig.ID))
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
	server("node-1.yaml")
	time.Sleep(2 * time.Second)
}
