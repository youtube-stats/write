package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

func handleConnection(c net.Conn) {
	defer func() {
		err := c.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(4)
		}
	}()

	key := getRandomKey()
	fmt.Println("Serving", c.RemoteAddr().String(), key)
	{
		_, err := c.Write([]byte(key))
		{
			if err != nil {
				fmt.Println(err)
				os.Exit(4)
			}
		}
	}
}

type LengthType = uint8

const (
	port = "0.0.0.0:3333"
	sleepTime = 60
	google = "https://www.googleapis.com/youtube/v3/channels?part=id&id=UC-lHJZR3Gqxm24_Vd_AJ5Yw&key=%s"
)

var (
	keys []string
	status []bool
	length LengthType
)

func checkKey(key string) bool {
	url := fmt.Sprintf(google, key)
	resp, err := http.Head(url)
	if err != nil {
		fmt.Println(key, err)
		return false
	}

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

func keyAudit() {
	fmt.Println("Key audit started")

	for {
		for i := LengthType(0); i < length; i++ {
			time.Sleep(sleepTime * time.Second)
			status[i] = checkKey(keys[i])
		}
	}
}

func getRandomKey() string {
	goodKeys := make([]string, 0)
	for i := LengthType(0); i < length; i++ {
		if status[i] {
			goodKeys = append(goodKeys, keys[i])
		}
	}

	if len(goodKeys) == 0 {
		fmt.Println("No good keys right now")
		os.Exit(6)
	}

	n := rand.Intn(len(goodKeys))

	return goodKeys[n]
}

func init() {
	fmt.Println("Key service started")
	rand.Seed(time.Now().Unix())

	keys = os.Args[1:]
	length = LengthType(len(keys))
	{
		if length == 0 {
			fmt.Println("No keys")
			os.Exit(1)
		}

		fmt.Println("Received", keys)
	}

	status = make([]bool, length)
	{
		for i := LengthType(0); i < length; i++ {
			status[i] = checkKey(keys[i])
		}
	}
}

func main() {
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
	go keyAudit()

	for {
		fmt.Println("Waiting for connection")
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
