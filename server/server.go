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
	// BUFFER_SIZE The node's buffer size
	BUFFER_SIZE=4096

	// RANDOM_PULL_FREQUENCY defines how frequently a node will pull
	// the state from a random node
	RANDOM_PULL_FREQUENCY = 3 * time.Second

	PULL = "pull"
)

// NodeContext contains all the required info of a server in the network
type NodeContext struct {
	hostname string
	port string
	data *Pair
	Nodes map[string]*Pair
	blacklist map[string]bool
	malicious *bool

	nodesMutex sync.Mutex
	blackListMutex sync.Mutex
}

// CreateNodeContext Creates a new NodeContext
func CreateNodeContext(hostname string, port string) NodeContext {
	return NodeContext{hostname: hostname, port: port, data:new(Pair), Nodes:make(map[string]*Pair),
		blacklist:make(map[string]bool), malicious: new(bool)}
}

func (nodeCtx *NodeContext) GetData() *Pair {
	return &(*nodeCtx.data)
}

func (ctx NodeContext) isMalicious() bool {
	return *ctx.malicious
}

func (ctx NodeContext) SetMalicious() {
	*ctx.malicious = true
}

func (ctx NodeContext) IsBlackListed(address string) bool {
	if _, exists := ctx.blacklist[address]; exists && ctx.blacklist[address] {
		return true
	}
	return false
}

func (ctx NodeContext) AddToBlackList(address string) {
	ctx.blacklist[address] = true
}

// ListNodes prints all the nodes of the given NodeContext instance to stdout
func ListNodes(nodeCtx NodeContext, debug bool) {
	fmt.Printf("%s:%s --> %d", nodeCtx.hostname, nodeCtx.port, nodeCtx.GetData().GetData())
	if debug {
		fmt.Printf(", %d", nodeCtx.GetData().GetTs())
	}
	fmt.Println()
	for addr, data := range nodeCtx.Nodes {
		fmt.Printf("%s --> %d", addr, data.GetData())
		if debug {
			fmt.Printf(", %d", data.GetTs())
		}
		fmt.Println()
	}
}

// The ConnectionHandler runs as an independent process that waits for input connections
// In case of a pull request (the message starts with the substring "pull"), it sends its state
// to the requester.
func ConnectionHandler(conn net.Listener, nodeCtx NodeContext) NodeContext {
	fmt.Printf("\nTCP socket started\n>>>")
	for true {
		response, err := conn.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		buf := make([]byte, BUFFER_SIZE)

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

			// Ignore updates from the "future"
			if ts > time.Now().UnixNano() {
				continue
			}

			if err != nil {
				fmt.Println(err)
			}

			address := ip + ":" +port
			if _, in := nodeCtx.Nodes[address] ; in {
				// Ignore updates with a ts older than one we already have
				if ts < nodeCtx.Nodes[address].GetTs() {
					continue
				}
				nodeCtx.Nodes[address].SetData(val).SetTs(ts)
			} else {
				nodeCtx.Nodes[address] = CreatePair(val, ts)
			}
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
		nodeCtx.blacklist[dst_addr] = true
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
		nodeCtx.blacklist[dst_addr] = true
		delete(nodeCtx.Nodes, dst_addr)
	} else {
		var ts int64

		if nodeCtx.isMalicious() {
			ts = 0
		} else {
			ts = nodeCtx.GetData().GetTs()
		}

		outString := fmt.Sprintf("%s:%s,%d,%d\n", nodeCtx.hostname, nodeCtx.port, nodeCtx.GetData().GetData(),
			ts)
		for address, data := range nodeCtx.Nodes {
			if nodeCtx.isMalicious() {
				ts = time.Date(20100, 1, 1, 1, 1, 1, 1, time.Local).UnixNano()
			} else {
				ts = nodeCtx.GetData().GetTs()
			}
			str := fmt.Sprintf("%s,%d,%d\n", address, data.GetData(), ts)
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

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}