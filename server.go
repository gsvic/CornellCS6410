package main

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

    LISTEN = "listen"
    ADD_NODES = '+'
    LIST_NODES = '?'

    HI = "hi"
    END_STR = "?EOF?"
)

type State struct {
  data int
}

func main() {
    reader := bufio.NewReader(os.Stdin)

    fmt.Printf("Enter the port number: ")
    port, _ := reader.ReadString('\n')
    port = port[:len(port)-1]
    nodes := make(map[string]int)

    conn, err := net.Listen(CONN_TYPE, CONN_HOST+":"+port)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }

    fmt.Printf("Running node at %s:%s\n", CONN_HOST, port)

    go Socket(conn, nodes)

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
        nodes[input[1:]] = -1

        /* Say hi to the new node */
        ReportChange(CONN_HOST, port, "hi"+END_STR, ip+":"+port)

      } else if input[0] == LIST_NODES {
        fmt.Println(nodes)
      } else if data, err := strconv.Atoi(input); err == nil {
        fmt.Printf("%s:%s --> %d\n", CONN_HOST, port, data)

      // Send to the rest of the nodes
      for address, _ := range nodes {
        ReportChange(CONN_HOST, port, input, address)
      }
  }

      // Check if exit
      if strings.ToLower(input) == "exit" {
        conn.Close()
        fmt.Println("Bye!")
        return;
      }


    }
}

func ReportChange(src_host string, src_port string, data string,
  dst_addr string) {

    ln, err := net.Dial("tcp", dst_addr)

    if err != nil {
      fmt.Println("error connecting to "+dst_addr)
      fmt.Println(err)
    } else {
      fmt.Printf("Sending: %s to %s of len: %d\n", data, dst_addr, data)
      fmt.Fprintf(ln, "%s:%s:%s:",src_host,src_port,data)
    }

}

// Start socket
func Socket(conn net.Listener, nodes map[string]int) {
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
    split := strings.Split(string(buf), END_STR)

    // A new node said hi
    if split[0] == HI {
      ip := split[1]
      port := split[2]
      fmt.Printf("Node %s:%s says hi!\n>>>", ip, port)
    } else {
      ip := split[0]
      port := split[1]
      value := split[2]
      //remote := response.RemoteAddr()

      i64, err := strconv.Atoi(value)

      if err != nil {
        fmt.Println(err)
      }

      nodes[ip+":"+port] = int(i64)
    }
  }
}
