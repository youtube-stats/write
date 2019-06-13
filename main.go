package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}

		result := "hello\n"
		c.Write([]byte(string(result)))
	}
	c.Close()
}

const (
	port = "0.0.0.0:3333"
	google = "https://www.googleapis.com/youtube/v3/channels?part=statistics&key=UC-lHJZR3Gqxm24_Vd_AJ5Yw&id=%s"
)

func checkKey(key string) bool {
	url := fmt.Sprintf(google, key)
	resp, err := http.Head(url)
	{
		if err != nil {
			fmt.Println(key, err)
			return false
		}
	}

	fmt.Println()
	status := resp.StatusCode == http.StatusOK
	{
		if status {
			fmt.Println(key, "Good")
		} else {
			fmt.Println(key, "Bad")
		}
	}

	return status
}

func init() {
	rand.Seed(time.Now().Unix())
	fmt.Println("Key service started")
}

func main() {
	args := os.Args[1:]
	{
		if len(args) == 0 {
			fmt.Println("No keys")
			os.Exit(1)
		}

		fmt.Println("Received", args)
	}

	keys := make([]string, 0)
	{
		for _, key := range args {
			if checkKey(key) {
				keys = append(keys, key)
			}
		}
	}

	server, err := net.Listen("tcp4", port)
	{
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	defer func() {
		_ = server.Close()
	}()

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println(err)
			fmt.Println("Closing server")
			_ = server.Close()
			os.Exit(3)
		}
		go handleConnection(connection)
	}
}
