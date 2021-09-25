package main

import (
	"cs6410/gossip/server"
	"time"
)

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	CONN_TYPE = "tcp"

	DATA     = 'd'
	ADD_NODES  = '+'
	LIST_NODES = '?'
	LIST_NODES_DEBUG = '!'

	// ENABLE_MALICIOUS_MODE When this is enabled, the node will start sharing either
	// very old updates (01/01/1970) or updates from the future (01/01/2100)
	ENABLE_MALICIOUS_MODE = 'm'
)

func main() {
	// Parse command-line argumens
	port := flag.Int("port", server.RandNum(1024, 49151), "Port Number")
	localMode := flag.Bool("local", false, "Run in local mode (on localhost)")
	flag.Parse()

	var ip string
	if *localMode {
		ip = "localhost"
	} else {
		ip = server.GetLocalIP()
	}

	reader := bufio.NewReader(os.Stdin)

	strPort := strconv.Itoa(*port)
	conn, err := net.Listen(CONN_TYPE, ip+":"+strPort)
	for err != nil {
		strPort = strconv.Itoa(server.RandNum(1024, 49151))
		conn, err = net.Listen(CONN_TYPE, ip+":"+strPort)
	}

	// Initialize the Node Context
	nodeCtx := server.CreateNodeContext(ip, strPort)

	fmt.Printf("Running node at %s:%s\n", ip, strPort)

	// Start the ConnectionHandler
	nodeCtx.StartConnectionHandler(conn)
	nodeCtx.StartRandomPull()

	for true {
		fmt.Print(">>>")
		input, _ := reader.ReadString('\n')

		if len(input) == 1 {
			continue
		}

		input = input[:len(input)-1]

		if input[0] == ADD_NODES {
			split := strings.Split(input[1:], ":")
			ip := split[0]
			port := split[1]

			addr := ip + ":" + port

			if nodeCtx.IsBlackListed(addr) {
				fmt.Printf("Oops! Node %s is in blacklisted\n", addr)
				continue
			}

			nodeCtx.UpdateNode(addr, 0, 0)

			// Fetch the new node's state
			server.SendPullRequest(nodeCtx, ip+":"+port)

		} else if input[0] == LIST_NODES {
			server.ListNodes(nodeCtx, false)
		} else if input[0] == LIST_NODES_DEBUG {
			server.ListNodes(nodeCtx, true)
		} else if input[0] == ENABLE_MALICIOUS_MODE {
			if nodeCtx.IsMalicious() {
				nodeCtx.TurnMaliciousOff()
				fmt.Println("Adversarial Mode: OFF")
			} else {
				nodeCtx.TurnMaliciousOn()
				fmt.Println("Adversarial Mode: ON")
			}
		} else if data, err := strconv.Atoi(input); err == nil {
			nodeCtx.GetData().SetData(data).SetTs(time.Now().Unix())
			fmt.Printf("%s:%s --> %d\n", ip, strPort, nodeCtx.GetData().GetData())
		} else if input[0] ==  DATA {
			fmt.Printf("My Data -> %d\n", nodeCtx.GetData().GetData())
		} else if strings.ToLower(input) == "exit" {
			conn.Close()
			fmt.Println("Bye!")
			return
		} else {
			fmt.Printf("Invalid argument: %s\n", input)
		}
	}
}