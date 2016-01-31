package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var addr = flag.String("addr", ":3000", "Network address to listen on.")

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(2)
}

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		die(err)
	}

	defer l.Close()
	log.Println("listening on", l.Addr())
	for {
		conn, err := l.Accept()
		if err != nil {
			die(err)
		}

		log.Printf("received message %s -> %s", conn.RemoteAddr(), conn.LocalAddr())

		go io.Copy(os.Stdout, conn)
	}
}
