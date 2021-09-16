package main

import (
    "net"
    "fmt"
    "os"
    "bufio"
    "strings"
    "strconv"
  )

const (
    CONN_HOST = "localhost"
    CONN_TYPE = "tcp"
)

func main() {
    reader := bufio.NewReader(os.Stdin)
    port, _ := reader.ReadString('\n')
    port = port[:len(port)-1]
    nodes := make(map[string]int)
    //
    // // _, err := net.Listen(CONN_TYPE, CONN_HOST+":"+port)
    // // if err != nil {
    // //     fmt.Println("Error listening:", err.Error())
    // //     os.Exit(1)
    // // }

    // fmt.Println("Listening on " + CONN_HOST + ":" + port)
    // for {
    //     // Listen for an incoming connection.
    //     conn, err := l.Accept()
    //     if err != nil {
    //         fmt.Println("Error accepting: ", err.Error())
    //         os.Exit(1)
    //     }
    //     // Handle connections in a new goroutine.
    //     go handleRequest(conn)
    // }

    input := ""
    for input != "exit" {
      fmt.Print(">>>")
      input, _ := reader.ReadString('\n')
      input = input[:len(input)-1]
      fmt.Printf("Got: '%s'\n", input)

      if input[0] == '+' {
        split := strings.Split(input[1:], ":")
        ip := split[0]
        port := split[1]
        fmt.Printf("Node added[ip=%s, port=%s]\n", ip, port)
        nodes[input[1:]] = -1
      } else if input[0] == '?' {
        fmt.Println(nodes)
      } else {
      data, _ := strconv.Atoi(input)
      fmt.Printf("%s:%s --> %d\n", CONN_HOST, port, data)
    }

      // Check if exit
      if strings.ToLower(input) == "exit" {
        fmt.Println("Bye!")
        return;
      }


    }
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
  // Make a buffer to hold incoming data.
  buf := make([]byte, 1024)
  // Read the incoming connection into the buffer.
  _, err := conn.Read(buf)
  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }
  fmt.Println("Got: ", string(buf))
  // Send a response back to person contacting us.
  conn.Write([]byte("Message received."))
  // Close the connection when you're done with it.
  conn.Close()
}
