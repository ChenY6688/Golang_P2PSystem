package main

import (
	"net/http"
	"poker/p2p"
	"time"
)

func makeServerAndStart(addr, apiAddr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		Version:       "POKER V0.1-alpha",
		ListenAddr:    addr,
		APIListenAddr: apiAddr,
		GameVariant:   p2p.TexasHoldem,
	}
	server := p2p.NewServer(cfg)
	go server.Start()

	time.Sleep(time.Millisecond * 200)

	return server
}

func main() {
	// node1 := makeServerAndStart(":3000", ":3001")
	// node2 := makeServerAndStart(":4000", ":4001")
	// node3 := makeServerAndStart(":5000", ":5001")
	// node4 := makeServerAndStart(":6000", ":6001")

	// node2.Connect(node1.ListenAddr)
	// node3.Connect(node2.ListenAddr)
	// node4.Connect(node3.ListenAddr)

	// go func() {
	// 	time.Sleep(time.Second * 2)
	// 	http.Get("http://localhost:3001/takeseat")
	// 	time.Sleep(time.Second * 2)
	// 	http.Get("http://localhost:4001/takeseat")
	// 	time.Sleep(time.Second * 2)
	// 	http.Get("http://localhost:5001/takeseat")

	// }()

	// select {}
	// return
	playerB := makeServerAndStart(":4000", ":4001")
	playerC := makeServerAndStart(":5000", ":5001")
	playerD := makeServerAndStart(":7000", ":7001")
	playerA := makeServerAndStart(":8000", ":8001")

	go func() {

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/takeseat")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:5001/takeseat")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:7001/takeseat")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:8001/takeseat")

		// PREFLOP
		time.Sleep(time.Second * 10)
		http.Get("http://localhost:5001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:7001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:8001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/fold")

		// // FLOP
		time.Sleep(time.Second * 2)
		http.Get("http://localhost:5001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:7001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:8001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/fold")

		// TURN
		time.Sleep(time.Second * 2)
		http.Get("http://localhost:5001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:7001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:8001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/fold")

		// RIVER
		time.Sleep(time.Second * 2)
		http.Get("http://localhost:5001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:7001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:8001/fold")

		time.Sleep(time.Second * 2)
		http.Get("http://localhost:4001/fold")

	}()

	time.Sleep(time.Millisecond * 200)
	playerC.Connect(playerB.ListenAddr)

	time.Sleep(time.Millisecond * 200)
	playerD.Connect(playerC.ListenAddr)

	time.Sleep(time.Millisecond * 200)
	playerA.Connect(playerD.ListenAddr)

	select {}
}
