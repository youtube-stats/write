package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type ChannelRow struct {
	id int32
	serial [24]byte
}

func handleConnection(c net.Conn) {
	defer func() {
		err := c.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(4)
		}
	}()

	bytes := make([]byte, 4)
	n, err := c.Read(bytes)
	if err != nil {
		fmt.Println(err)
		os.Exit(7)
	}

	fmt.Println("Retrieved", n, "bytes")
	{
		message := "Hello\n"
		_, err := c.Write([]byte(message))
		{
			if err != nil {
				fmt.Println(err)
				os.Exit(4)
			}
		}
	}
}

const (
	port = "0.0.0.0:3334"
	sleepTime = 3600
	sqlUrl = "postgresql://admin@localhost:5432/youtube?sslmode=disable"
	sqlQuery = "SELECT id, serial FROM youtube.stats.channels ORDER BY id ASC"
)

var (
	rows []ChannelRow
)

func setChannels() {
	fmt.Println("Updating channels")

	db, err := sql.Open("postgres", sqlUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(5)
	}

	defer func() {
		fmt.Println("Closing sql connection")
		_ = db.Close()
	}()

	results, err := db.Query(sqlQuery)
	if err != nil {
		fmt.Println(err)
		os.Exit(6)
	}

	tmp := make([]ChannelRow, 0)
	var row ChannelRow

	for results.Next() {
		err := results.Scan(row)
		if err != nil {
			fmt.Println(err)
			os.Exit(7)
		}

		tmp = append(tmp, row)
	}

	fmt.Println("Retrieved", len(tmp), "channels")
	rows = tmp
}

func channelUpdate() {
	for {
		fmt.Println("Waiting for", sleepTime, "seconds")
		time.Sleep(sleepTime * time.Second)
		setChannels()
	}
}

func init() {
	fmt.Println("Cache service started")
	rand.Seed(time.Now().Unix())

	setChannels()
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

	go channelUpdate()

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
