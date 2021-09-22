package main

import (
	"cs6410/gossip/server"
	"time"
)

import (
	"bufio"
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

	HI      = "hi"
	END_STR = ":"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter the port number: ")
	port, _ := reader.ReadString('\n')
	port = port[:len(port)-1]

	nodeCtx := server.New(CONN_HOST, port)

	conn, err := net.Listen(CONN_TYPE, CONN_HOST+":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Running node at %s:%s\n", CONN_HOST, port)

	go server.Socket(conn, nodeCtx)

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
			fmt.Printf("Node added[ip=%s, port=%s]\n", ip, port)
			//nodeCtx.nodes[input[1:]] = -1

			/* Say hi to the new node */
			server.ReportState(nodeCtx, ip+":"+port)

		} else if input[0] == LIST_NODES {
			server.ListNodes(nodeCtx)
		} else if data, err := strconv.Atoi(input); err == nil {
			(*nodeCtx.Data).Data = data
			(*nodeCtx.Data).Ts = time.Now().Unix()
			fmt.Printf("%s:%s --> %d\n", CONN_HOST, port, server.GetData(nodeCtx))
			// Send to the rest of the nodes
			for address, _ := range server.GetNodeMap(nodeCtx) {
				server.ReportState(nodeCtx, address)
			}
		} else if input[0] ==  DATA {
			fmt.Printf("My Data -> %d\n", server.GetData(nodeCtx))
		}

		// Check if exit
		if strings.ToLower(input) == "exit" {
			conn.Close()
			fmt.Println("Bye!")
			return
		}

	}
}