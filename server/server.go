package server

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type NodeContext struct {
	hostname string
	port string
	Data *Pair
	Nodes map[string]Pair
}

type Pair struct {
	data int
	ts int64
}

func CreatePair(data int, ts int64) Pair {
	return Pair{data, ts}
}

func New(hostname string, port string) NodeContext {
	return NodeContext{hostname, port, new(Pair), make(map[string]Pair)}
}

func ListNodes(nodeCtx NodeContext) {
	fmt.Println(nodeCtx.Nodes)
}

//func SetData(nodeCtx NodeContext, data int) {
//	*nodeCtx.Data = data
//}

func GetData(nodeCtx NodeContext) Pair {
	return *nodeCtx.Data
}

func GetNodeMap(ctx NodeContext) map[string]Pair {
	return ctx.Nodes
}

// Start socket
func Socket(conn net.Listener, state NodeContext) NodeContext {
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
		split := strings.Split(string(buf), ":")

		ip := split[0]
		port := split[1]
		value := split[2]
		//remote := response.RemoteAddr()

		// A new node said hi
		if value == "hi" {
			fmt.Printf("Node %s:%s says hi! Let's update it with %d\n>>>", ip, port, state.Data)
			SendMsg(state, strconv.Itoa((*state.Data).data), ip+":"+port)
		} else {
			i64, err := strconv.Atoi(value)

			if err != nil {
				fmt.Println(err)
			}

			fmt.Printf("Updating node %s:%s value with %d\n>>>", ip, port, i64)
			state.Nodes[ip+":"+port] = Pair{i64, time.Now().Unix()}
		}
	}
	return state
}



func SendMsg(nodeCtx NodeContext, msg string, dst_addr string) {
	ln, err := net.Dial("tcp", dst_addr)

	if err != nil {
		fmt.Println("error connecting to " + dst_addr)
		fmt.Println(err)
	} else {
		fmt.Fprintf(ln, "%s:%s:%s:", nodeCtx.hostname, nodeCtx.port, msg)
	}

}
