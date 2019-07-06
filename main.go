package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/sparrc/go-ping"
)

func main() {
	var host, statsdAddress string
	var sleep time.Duration

	flag.StringVar(&host, "host", "www.google.com", "host to check")
	flag.StringVar(&statsdAddress, "statsd", "127.0.0.1:8125", "udp statsd address")
	flag.DurationVar(&sleep, "sleep", 2*time.Minute, "sleep duration between checks")
	flag.Parse()

	conn, err := open(statsdAddress)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		pinger, err := ping.NewPinger("www.google.com")
		if err != nil {
			log.Printf("FAILED %v", err)
			conn.Printf("pingcheck.run.failed:1|g\n")
		} else {
			pinger.Count = 3
			pinger.SetPrivileged(true)
			pinger.Run()
			s := pinger.Statistics()

			log.Printf("OK loss: %.3f rtt: %v", s.PacketLoss, s.AvgRtt)
			conn.Printf("pingcheck.run.ok:1|g\n")
			conn.Printf("pingcheck.packet_loss:%.3f|g\n", s.PacketLoss)
			conn.Printf("pingcheck.rtt:%d|ms\n", s.AvgRtt/time.Millisecond)
		}

		time.Sleep(sleep)
	}
}

type connection struct {
	addr *net.UDPAddr
	conn *net.UDPConn
}

func open(writeAddress string) (*connection, error) {
	addr, err := net.ResolveUDPAddr("udp", writeAddress)
	if err != nil {
		return nil, fmt.Errorf("Couldn't resolve %q, %v", writeAddress, err)
	}

	bind, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, fmt.Errorf("Couldn't resolve :0, %v", err)
	}

	conn, err := net.ListenUDP("udp", bind)
	if err != nil {
		return nil, fmt.Errorf("Couldn't listen %s, %v", addr, err)
	}

	return &connection{
		addr: addr,
		conn: conn,
	}, nil
}

func (c *connection) Printf(tpl string, v ...interface{}) {
	c.conn.WriteToUDP([]byte(fmt.Sprintf(tpl, v...)), c.addr)
}
