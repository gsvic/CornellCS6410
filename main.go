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

	HI      = "hi"
	END_STR = ":"
)

func main() {
	port := flag.Int("port", server.RandNum(1024, 49151), "Port Number")
	flag.Parse()

	strPort := strconv.Itoa(*port)

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter the port number: ")

	nodeCtx := server.CreateNodeContext(CONN_HOST, strPort)

	conn, err := net.Listen(CONN_TYPE, CONN_HOST+":"+strPort)
	for err != nil {
		strPort = strconv.Itoa(server.RandNum(1024, 49151))
		conn, err = net.Listen(CONN_TYPE, CONN_HOST+":"+strPort)
	}

	fmt.Printf("Running node at %s:%s\n", CONN_HOST, strPort)

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
			fmt.Printf("Node added[ip=%s, port=%s]\n", ip, port)
			nodeCtx.Nodes[input[1:]] = &server.Pair{0, 0}

			/* Say hi to the new node */
			//server.ReportState(nodeCtx, ip+":"+port)
			server.SendPullRequest(nodeCtx, ip+":"+port)

			if !randomPullStarted {
				go server.RandomPull(nodeCtx)
				randomPullStarted = true
			}

		} else if input[0] == LIST_NODES {
			server.ListNodes(nodeCtx, false)
		} else if input[0] == LIST_NODES_DEBUG {
			server.ListNodes(nodeCtx, true)
		} else if data, err := strconv.Atoi(input); err == nil {
			(*nodeCtx.Data).Data = data
			(*nodeCtx.Data).Ts = time.Now().Unix()
			fmt.Printf("%s:%s --> %d\n", CONN_HOST, strPort, (*nodeCtx.Data).Data)
		} else if input[0] ==  DATA {
			fmt.Printf("My Data -> %d\n", (*nodeCtx.Data).Data)
		}

		// Check if exit
		if strings.ToLower(input) == "exit" {
			conn.Close()
			fmt.Println("Bye!")
			return
		}

	}
}