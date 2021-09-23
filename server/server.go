package server

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PULL = "pull"
	RANDOM_PULL_FREQUENCY = 3 * time.Second
)

type NodeContext struct {
	hostname string
	port string
	Data *Pair
	Nodes map[string]*Pair
	BlackList map[string]bool
	ismalicious *bool

	nodesMutex sync.Mutex
	blackListMutex sync.Mutex
}

// CreateNodeContext Creates a new NodeContext
func CreateNodeContext(hostname string, port string) NodeContext {
	return NodeContext{hostname: hostname, port: port, Data:new(Pair), Nodes:make(map[string]*Pair),
		BlackList:make(map[string]bool), ismalicious: new(bool)}
}
func (ctx *NodeContext) Yo() {

}

// GetValue Get the value of a NodeContext
func GetValue(ctx NodeContext) int {
	return (*ctx.Data).Data
}

// GetTs Get the timestamp of a NodeContext
func GetTs(ctx NodeContext) int64 {
	return (*ctx.Data).Ts
}

func (ctx NodeContext) isMalicious() bool {
	return *ctx.ismalicious
}

func (ctx NodeContext) SetMalicious() {
	*ctx.ismalicious = true
}

// ListNodes prints all the nodes of the given NodeContext instance to stdout
func ListNodes(nodeCtx NodeContext, debug bool) {
	fmt.Printf("%s:%s --> %d", nodeCtx.hostname, nodeCtx.port, (*nodeCtx.Data).Data)
	if debug {
		fmt.Printf(", %d", (*nodeCtx.Data).Ts)
	}
	fmt.Println()
	for addr, data := range nodeCtx.Nodes {
		fmt.Printf("%s --> %d", addr, data.Data)
		if debug {
			fmt.Printf(", %d", data.Ts)
		}
		fmt.Println()
	}
}

// The ConnectionHandler runs as an independent process that waits for input connections
// In case of a pull request (the message starts with the substring "pull"), it sends its state
// to the requester.
func ConnectionHandler(conn net.Listener, nodeCtx NodeContext) NodeContext {
	fmt.Printf("\nTCP socket started\n")
	for true {
		response, err := conn.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		buf := make([]byte, 1024)

		_, err = response.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
		}

		strBuf := string(buf)

		if strings.Contains(strBuf, PULL) {
			split := strings.Split(strBuf, ":")
			ip := split[1]
			port := strings.Split(split[2], "\n")[0]
			ReportState(nodeCtx, fmt.Sprintf("%s:%s", ip, port))

			continue
		}

		nodes := strings.Split(string(buf), "\n")
		nodes = nodes[:len(nodes)-1]

		for _, node := range nodes {
			csv := strings.Split(node, ",")

			ip := strings.Split(csv[0], ":")[0]
			port := strings.Split(csv[0], ":")[1]

			if ip == nodeCtx.hostname && port == nodeCtx.port {
				continue
			}

			value := csv[1]
			timestamp := csv[2]

			val, err := strconv.Atoi(value)

			if err != nil {
				fmt.Println(err)
			}

			ts, err := strconv.ParseInt(timestamp, 10, 64)

			if err != nil {
				fmt.Println(err)
			}

			//fmt.Printf("Updating node %s:%s value with %d\n>>>", ip, port, i64)
			nodeCtx.Nodes[ip+":"+port] = &Pair{val, ts}
		}
	}
	return nodeCtx
}

// SendPullRequest sends a pull request to the destination address
func SendPullRequest(nodeCtx NodeContext, dst_addr string) {
	ln, err := net.Dial("tcp", dst_addr)

	if err != nil {
		fmt.Println("error connecting to " + dst_addr)
		fmt.Println(err)

		// Put the address to the blacklist in case of failure
		nodeCtx.BlackList[dst_addr] = true
		delete(nodeCtx.Nodes, dst_addr)
	} else {
		fmt.Fprintf(ln, "pull:%s:%s\n", nodeCtx.hostname, nodeCtx.port)
	}

}

// ReportState reports the state of the given NodeContext to the destination node
func ReportState(nodeCtx NodeContext, dst_addr string) {
	ln, err := net.Dial("tcp", dst_addr)

	if err != nil {
		fmt.Println("error connecting to " + dst_addr)
		fmt.Println(err)

		// Put the address to the blacklist in case of failure
		nodeCtx.BlackList[dst_addr] = true
		delete(nodeCtx.Nodes, dst_addr)
	} else {
		var ts int64

		if nodeCtx.isMalicious() {
			ts = 0
		} else {
			ts = GetTs(nodeCtx)
		}

		outString := fmt.Sprintf("%s:%s,%d,%d\n", nodeCtx.hostname, nodeCtx.port, GetValue(nodeCtx),
			ts)
		for address, data := range nodeCtx.Nodes {
			if nodeCtx.isMalicious() {
				ts = 0
			} else {
				ts = time.Date(20100, 1, 1, 1, 1, 1, 1, time.Local).UnixNano()
			}
			str := fmt.Sprintf("%s,%d,%d\n", address, data.Data, ts)
			outString = outString + str
		}

		fmt.Fprintf(ln, outString)
	}
}

// RandomPull sends a random pull request
func RandomPull(nodeCtx NodeContext) {
	for true {
		if len(nodeCtx.Nodes) > 0 {
			addresses := reflect.ValueOf(nodeCtx.Nodes).MapKeys()
			randomIdx := rand.Intn(len(addresses))
			randomAddress := addresses[randomIdx]

			SendPullRequest(nodeCtx, randomAddress.String())
		}
		time.Sleep(RANDOM_PULL_FREQUENCY)
	}

}

// RandNum A simple random number generator
func RandNum(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max - min + 1) + min
}