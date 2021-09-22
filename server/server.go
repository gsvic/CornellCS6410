package server

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type NodeContext struct {
	hostname string
	port string
	Data *Pair
	Nodes map[string]Pair
}

type Pair struct {
	Data int
	Ts int64
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
		csv := strings.Split(strings.Split(string(buf), "\n")[0], ",")

		ip := strings.Split(csv[0],":")[0]
		port := strings.Split(csv[0],":")[1]
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
		if _, ok := state.Nodes[ip+":"+port]; ok {
			state.Nodes[ip+":"+port] = Pair{val, ts}
		} else {
			state.Nodes[ip+":"+port] = Pair{val, ts}
			ReportState(state, ip+":"+port)
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

func ReportState(nodeCtx NodeContext, dst_addr string) {
	ln, err := net.Dial("tcp", dst_addr)

	if err != nil {
		fmt.Println("error connecting to " + dst_addr)
		fmt.Println(err)
	} else {
		fmt.Fprintf(ln, "%s:%s,%d,%d\n", nodeCtx.hostname, nodeCtx.port, nodeCtx.Data.Data, nodeCtx.Data.Ts)
	}

}