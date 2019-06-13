package main

import (
	"./message"
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type ChannelRow struct {
	time time.Time
	id int32
	sub int32
}

const (
	port = "0.0.0.0:3335"
	sqlUrl = "postgresql://admin@localhost:5432/youtube?sslmode=disable"
	sqlInsert = "INSERT INTO youtube.stats.channels (time, id, subs) VALUE ($1, $2, $3)"
	cacheSize = 1000
)

var (
	cache []ChannelRow
)

func handleConnection(c net.Conn) {
	var bytes []byte
	{
		{
			n, err := c.Read(bytes)
			if err != nil {
				fmt.Println(err)
				os.Exit(7)
			}

			fmt.Println("Retrieved", n, "bytes")
		}

		{
			err := c.Close()
			if err != nil {
				fmt.Println(err)
				os.Exit(4)
			}
		}
	}

	msg := &message.ChannelMessage{}
	{
		err := proto.Unmarshal(bytes, msg)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	for i := 0; i < len(msg.Ids); i++ {
		time := time.Now()
		id := msg.Ids[i]
		sub := msg.Subs[i]

		row := ChannelRow{time, id, sub}
		cache = append(cache, row)
	}
}

func write() {
	fmt.Println("Writing", cacheSize, "channels")
	db, err := sql.Open("postgres", sqlUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(5)
	}

	defer func() {
		fmt.Println("Closing sql connection")
		_ = db.Close()
	}()

	txn, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		os.Exit(8)
	}

	defer func() {
		err = txn.Commit()
		if err != nil {
			fmt.Println(err)
			os.Exit(9)
		}
	}()

	for i := 0; i < cacheSize; i++ {
		row := cache[i]
		_, err := txn.Exec(sqlInsert, row.time, row.id, row.sub)
		if err != nil {
			fmt.Println(err)
			os.Exit(6)
		}
	}

	cache = cache[cacheSize:]
	fmt.Println("New size of cache", len(cache))
}

func init() {
	fmt.Println("Cache service started")
	rand.Seed(time.Now().Unix())
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
