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
	CONN_HOST = "localhost"
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
	port := flag.Int("port", server.RandNum(1024, 49151), "Port Number")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	strPort := strconv.Itoa(*port)
	conn, err := net.Listen(CONN_TYPE, CONN_HOST+":"+strPort)
	for err != nil {
		strPort = strconv.Itoa(server.RandNum(1024, 49151))
		conn, err = net.Listen(CONN_TYPE, CONN_HOST+":"+strPort)
	}

	// Initialize the Node Context
	nodeCtx := server.CreateNodeContext(CONN_HOST, strPort)

	nodeCtx.Yo()

	fmt.Printf("Running node at %s:%s\n", CONN_HOST, strPort)

	// Start the ConnectionHandler
	go server.ConnectionHandler(conn, nodeCtx)

	randomPullStarted := false

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

			if _, exists := nodeCtx.BlackList[addr]; exists && nodeCtx.BlackList[addr] {
				fmt.Printf("Oops! Node %s is in blacklisted\n", addr)
				continue
			}

			fmt.Printf("Node added[ip=%s, port=%s]\n", ip, port)
			nodeCtx.Nodes[input[1:]] = &server.Pair{0, 0}
			nodeCtx.BlackList[input[1:]] = false

			// Fetch the new node's state
			server.SendPullRequest(nodeCtx, ip+":"+port)

			// Start the random state pull if it has not started yet
			if !randomPullStarted {
				go server.RandomPull(nodeCtx)
				randomPullStarted = true
			}

		} else if input[0] == LIST_NODES {
			server.ListNodes(nodeCtx, false)
		} else if input[0] == LIST_NODES_DEBUG {
			server.ListNodes(nodeCtx, true)
		} else if input[0] == ENABLE_MALICIOUS_MODE {
			nodeCtx.SetMalicious()
		} else if data, err := strconv.Atoi(input); err == nil {
			(*nodeCtx.Data).Data = data
			(*nodeCtx.Data).Ts = time.Now().Unix()
			fmt.Printf("%s:%s --> %d\n", CONN_HOST, strPort, (*nodeCtx.Data).Data)
		} else if input[0] ==  DATA {
			fmt.Printf("My Data -> %d\n", (*nodeCtx.Data).Data)
		}

		if strings.ToLower(input) == "exit" {
			conn.Close()
			fmt.Println("Bye!")
			return
		}

	}
}