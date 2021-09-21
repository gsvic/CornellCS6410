package main

import (
    "net"
    "fmt"
    "os"
    "bufio"
    "strings"
    "strconv"
    "time"
  )

const (
    CONN_HOST = "localhost"
    CONN_TYPE = "tcp"

    LISTEN = "listen"
    ADD_NODES = '+'
    LIST_NODES = '?'
)

func main() {
    reader := bufio.NewReader(os.Stdin)

    fmt.Printf("Enter the port of the node: ")
    port, _ := reader.ReadString('\n')
    port = port[:len(port)-1]
    nodes := make(map[string]int)

    conn, err := net.Listen(CONN_TYPE, CONN_HOST+":"+port)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    fmt.Printf("Running node at %s:%s\n", CONN_HOST, port)

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

    for true {
      fmt.Print(">>>")
      input, _ := reader.ReadString('\n')

      if len(input) == 1 {
        continue
      }

      input = input[:len(input)-1]
      //fmt.Printf("CMD: '%s'\n", input)

      if input[0] == ADD_NODES {
        split := strings.Split(input[1:], ":")
        ip := split[0]
        port := split[1]
        fmt.Printf("Node added[ip=%s, port=%s]\n", ip, port)
        nodes[input[1:]] = -1
      } else if input[0] == LIST_NODES {
        fmt.Println(nodes)
      } else if strings.ToLower(input) == LISTEN {
        // fmt.Println("Waiting for message...")
        // response, err := conn.Accept()
        // fmt.Println("Accepted")
        // if err != nil {
        //     fmt.Println("Error accepting: ", err.Error())
        //     os.Exit(1)
        // } else {
        //   go handleRequest(response)
        //   fmt.Printf("Done\n")
        // }
        go handleRequest(conn)
      } else if input == "conc" {
        go sth()
      } else if data, err := strconv.Atoi(input); err == nil {
      //data, _ := strconv.Atoi(input)
      fmt.Printf("%s:%s --> %d\n", CONN_HOST, port, data)

      // Send to the rest of the nodes
      for address, _ := range nodes {
          fmt.Printf("Sending: %d to %s\n", data, address)
          ln, err := net.Dial("tcp", address)

          if err != nil {
            fmt.Println("error connecting to "+address)
            fmt.Println(err)
          } else {
            fmt.Fprintf(ln, "hello from: "+port)
          }
      }
  }

      // Check if exit
      if strings.ToLower(input) == "exit" {
        fmt.Println("Bye!")
        return;
      }


    }
}

func sth() {
    for true {
      fmt.Println("\nI am running forever....")
      fmt.Printf(">>>")
      time.Sleep(5*time.Second)
    }
}

func sth2() {
    for true {
      fmt.Println("\nI am running forever....")
      fmt.Printf(">>>")
      time.Sleep(5*time.Second)
    }
}

// Handles incoming requests.
func handleRequest(conn net.Listener) {
  for true {
    response, err := conn.Accept()
    if err != nil {
        fmt.Println("Error accepting: ", err.Error())
        os.Exit(1)
    }


    // Make a buffer to hold incoming data.
    buf := make([]byte, 1024)
    // Read the incoming connection into the buffer.
    fmt.Println("Waiting for message...")
    _, err = response.Read(buf)
    if err != nil {
      fmt.Println("Error reading:", err.Error())
    }
    fmt.Printf("Received: %s\n>>>", string(buf))
    // Send a response back to person contacting us.
    //conn.Write([]byte("Message received."))
    // Close the connection when you're done with it.
    //conn.Close()
  }
}
