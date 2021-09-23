package server

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	PULL = "pull"
)

type NodeContext struct {
	hostname string
	port string
	Data *Pair
	Nodes map[string]*Pair
}

type Pair struct {
	Data int
	Ts int64
}

func CreateNodeContext(hostname string, port string) NodeContext {
	return NodeContext{hostname, port, new(Pair), make(map[string]*Pair)}
}

func GetValue(ctx NodeContext) int {
	return (*ctx.Data).Data
}

func GetTs(ctx NodeContext) int64 {
	return (*ctx.Data).Ts
}

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

func ConnectionHandler(conn net.Listener, nodeCtx NodeContext) NodeContext {
	fmt.Printf("\nTCP socket started\n>>>")
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

func SendPullRequest(nodeCtx NodeContext, dst_addr string) {
	ln, err := net.Dial("tcp", dst_addr)

	if err != nil {
		fmt.Println("error connecting to " + dst_addr)
		fmt.Println(err)
	} else {
		fmt.Fprintf(ln, "pull:%s:%s\n", nodeCtx.hostname, nodeCtx.port)
	}

}

func ReportState(nodeCtx NodeContext, dst_addr string) {
	ln, err := net.Dial("tcp", dst_addr)

	if err != nil {
		fmt.Println("error connecting to " + dst_addr)
		fmt.Println(err)
	} else {
		outString := fmt.Sprintf("%s:%s,%d,%d\n", nodeCtx.hostname, nodeCtx.port, GetValue(nodeCtx), GetTs(nodeCtx))
		for address, data := range nodeCtx.Nodes {
			str := fmt.Sprintf("%s,%d,%d\n", address, data.Data, data.Ts)
			outString = outString + str
		}

		fmt.Fprintf(ln, outString)
	}
}

func RandomPull(nodeCtx NodeContext) {
	for true {
		addresses := reflect.ValueOf(nodeCtx.Nodes).MapKeys()
		randomIdx := rand.Intn(len(addresses))
		randomAddress := addresses[randomIdx]
		SendPullRequest(nodeCtx, randomAddress.String())
		time.Sleep(5*time.Second)
	}

}

func RandNum(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max - min + 1) + min
}