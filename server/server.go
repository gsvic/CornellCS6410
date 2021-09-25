package server

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// BUFFER_SIZE The node's buffer size
	BUFFER_SIZE=4096
	// RANDOM_PULL_FREQUENCY defines how frequently a node will pull
	// the state from a random node
	RANDOM_PULL_FREQUENCY = 3 * time.Second
	MAX_IPS_PER_NODE = 3
	PULL = "pull"
)

// NodeContext contains all the required info of a server in the network
type NodeContext struct {
	hostname string
	port string
	data *Pair
	blacklist map[string]bool

	nodes map[string]map[string]*Pair

	malicious *bool
}

// CreateNodeContext Creates a new NodeContext
func CreateNodeContext(hostname string, port string) NodeContext {
	return NodeContext{hostname: hostname, port: port, data:new(Pair), blacklist:make(map[string]bool),
		malicious: new(bool), nodes: make(map[string]map[string]*Pair)}
}

// GetData returns node's data (value & timestamp)
func (nodeCtx *NodeContext) GetData() *Pair {
	return &(*nodeCtx.data)
}

// isMalicious tells if that node runs in malicious mode
func (nodeCtx NodeContext) IsMalicious() bool {
	return *nodeCtx.malicious
}

// TurnMaliciousOn turns malicious mode on
func (nodeCtx NodeContext) TurnMaliciousOn() {
	*nodeCtx.malicious = true
}

// TurnMaliciousOff turns malicious mode on
func (nodeCtx NodeContext) TurnMaliciousOff() {
	*nodeCtx.malicious = false
}

// IsBlackListed Checks if the given address is in the black list
func (nodeCtx NodeContext) IsBlackListed(address string) bool {
	if _, exists := nodeCtx.blacklist[address]; exists && nodeCtx.blacklist[address] {
		return true
	}
	return false
}

// AddToBlackList adds an address to the black list
func (nodeCtx NodeContext) AddToBlackList(address string) {
	nodeCtx.blacklist[address] = true
}

// UpdateNode update the node of the given address.
// If we see this node for a first time, we add it to the map
func (nodeCtx *NodeContext) UpdateNode(address string, data int, ts int64) bool {
	ip := strings.Split(address, ":")[0]
	port := strings.Split(address, ":")[1]

	reportChange := false

	// We have seen this ip
	if _, seenIP := nodeCtx.nodes[ip] ; seenIP {
		// We have seen this port - Just update this node
		if _, seenPort := nodeCtx.nodes[ip][port] ; seenPort {
			// Report the change
			if data != nodeCtx.nodes[ip][port].GetData() && ts != nodeCtx.nodes[ip][port].GetTs() {
				reportChange = true
				nodeCtx.nodes[ip][port].SetData(data).SetTs(ts)
			}
		} else {
			numPorts := len(nodeCtx.nodes[ip])
			if numPorts < MAX_IPS_PER_NODE {
				nodeCtx.nodes[ip][port] = CreatePair(data, ts)
				reportChange = true
			} else {
				// We reached the maximum number of ports for that IP
				// We need to keep the most recent ones

				// Find the oldest update
				var minTs int64 = math.MaxInt64
				var minPort string
				for p, data := range nodeCtx.nodes[ip] {
					if data.ts < minTs {
						minPort = p
						minTs = data.ts
					}
				}

				if ts > minTs {
					// Delete the node with the oldest update
					delete(nodeCtx.nodes[ip], minPort)
					// Add the new node
					nodeCtx.nodes[ip][port] = CreatePair(data, ts)
					reportChange = true
				}
			}
		}
	} else {
		// We haven't seen this IP, so let's add it
		nodeCtx.nodes[ip] = make(map[string]*Pair)
		nodeCtx.nodes[ip][port] = CreatePair(data, ts)
		reportChange = true
	}

	return reportChange
}

// ListNodes prints all the nodes of the given NodeContext instance to stdout
func ListNodes(nodeCtx NodeContext, debug bool) {
	fmt.Printf("%s:%s --> %d", nodeCtx.hostname, nodeCtx.port, nodeCtx.GetData().GetData())
	if debug {
		fmt.Printf(", %d", nodeCtx.GetData().GetTs())
	}
	fmt.Println()
	for ip, port := range nodeCtx.nodes {
		for p, data := range port {
			addr := ip + ":" +p
			fmt.Printf("%s --> %d", addr, data.GetData())
			if debug {
				fmt.Printf(", %d", data.GetTs())
			}
			fmt.Println()
		}
	}
}

// The connectionHandler runs as an independent process that waits for input connections
// In case of a pull request (the message starts with the substring "pull"), it sends its state
// to the requester.
func (nodeCtx *NodeContext) connectionHandler(conn net.Listener){
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
			continue
		}

		// Trim WhiteSpace
		strBuf := strings.TrimSpace(string(buf))

		pullPattern :=  `^pull:([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}|\w+):[0-9]{1,5}$`
		nodePattern := `^([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}|\w+):[0-9]{1,5},\d+,\d+?`

		// Check the first line
		split := strings.Split(strBuf, "\n")
		if len(split) < 1 {
			return
		}

		pullPatternMatch, _ := regexp.MatchString(pullPattern, split[0])
		nodePatternMatch, _ := regexp.MatchString(nodePattern, split[0])

		// We received a pull request
		if pullPatternMatch {
			s := strings.Split(split[0], ":")
			ip := s[1]
			port := strings.Split(s[2], "\n")[0]
			ReportState(*nodeCtx, fmt.Sprintf("%s:%s", ip, port))

			// No need to process further after we report the state
			continue
		}

		nodes := strings.Split(string(buf), "\n")
		//nodes = nodes[:256]

		for _, node := range nodes {
			nodePatternMatch, _ = regexp.MatchString(nodePattern, node)
			if !nodePatternMatch {
				continue
			}

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
				continue
			}

			// Ignore updates from the "future"
			if ts > time.Now().UnixNano() {
				continue
			}

			address := ip + ":" +port
			if nodeCtx.UpdateNode(address, val, ts) {
				fmt.Printf("\n%s --> %d\n>>>", address, val)
			}
		}
	}
}

// SendPullRequest sends a pull request to the destination address
func SendPullRequest(nodeCtx NodeContext, dst_addr string) {
	ln, err := net.Dial("tcp", dst_addr)

	if err != nil {
		fmt.Printf("error connecting to %s\n", dst_addr)

		s := strings.Split(dst_addr, ":")
		ip := s[0]
		port := s[1]

		// Put the address to the blacklist in case of failure
		nodeCtx.blacklist[dst_addr] = true
		delete(nodeCtx.nodes[ip], port)
	} else {
		fmt.Fprintf(ln, "pull:%s:%s\n", nodeCtx.hostname, nodeCtx.port)
	}

}

// ReportState reports the state of the given NodeContext to the destination node
func ReportState(nodeCtx NodeContext, dst_addr string) {
	ln, err := net.Dial("tcp", dst_addr)

	if err != nil {
		fmt.Printf("error connecting to %s\n", dst_addr)

		s := strings.Split(dst_addr, ":")
		ip := s[0]
		port := s[1]

		// Put the address to the blacklist in case of failure
		nodeCtx.blacklist[dst_addr] = true
		delete(nodeCtx.nodes[ip], port)
	} else {
		var ts int64

		var outString string
		if nodeCtx.IsMalicious() {
			ts = 0
			outString = fmt.Sprintf("dfadfadfasfdadfkadfhadkfadahjkfhjkahdjfahkjdfajh\n",
				nodeCtx.hostname, nodeCtx.port, nodeCtx.GetData().GetData(),
				ts)
		} else {
			ts = nodeCtx.GetData().GetTs()
			outString = fmt.Sprintf("%s:%s,%d,%d\n", nodeCtx.hostname, nodeCtx.port, nodeCtx.GetData().GetData(),
				ts)
		}

		for ip, port := range nodeCtx.nodes {
			for p, data := range port {
				if nodeCtx.IsMalicious() {
					ts = time.Date(20100, 1, 1, 1, 1, 1, 1, time.Local).UnixNano()
				} else {
					ts = nodeCtx.GetData().GetTs()
				}
				address := ip + ":" + p
				str := fmt.Sprintf("%s,%d,%d\n", address, data.GetData(), ts)
				outString = outString + str
			}
		}

		fmt.Fprintf(ln, outString)
	}
}

// RandomPull sends a random pull request
func (nodeCtx *NodeContext) randomPull() {
	for true {
		if len(nodeCtx.nodes) > 0 {
			ips := reflect.ValueOf(nodeCtx.nodes).MapKeys()
			randomIpIdx := rand.Intn(len(ips))
			randomIp := ips[randomIpIdx].String()

			ports := reflect.ValueOf(nodeCtx.nodes[randomIp]).MapKeys()
			if len(ports) > 0 {
				randomPortId := rand.Intn(len(ports))
				randomPort := ports[randomPortId].String()

				randomAddress := randomIp + ":" + randomPort

				SendPullRequest(*nodeCtx, randomAddress)
			}
		}
		time.Sleep(RANDOM_PULL_FREQUENCY)
	}

}

func (nodeCtx *NodeContext) StartRandomPull() {
	go nodeCtx.randomPull()
}

func (nodeCtx *NodeContext) StartConnectionHandler(listener net.Listener) {
	go nodeCtx.connectionHandler(listener)
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
	panic("Cannot get a routable IP for this machine.")
}