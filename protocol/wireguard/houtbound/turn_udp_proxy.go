// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT
//  _  _____                            
// (_)/ ____|                           
//  _| (___   ___  __ _  __ _ _ __ ___  
// | |\___ \ / _ \/ _` |/ _` | '__/ _ \ 
// | |____) |  __/ (_| | (_| | | | (_) | & Hiddify
// |_|_____/ \___|\__, |\__,_|_|  \___/ 
//                __/ |                
//                |___/                 
//
// Package main implements a TURN client with support for TCP
package houtbound

import (
	"fmt"
	"github.com/pion/logging"
	"github.com/pion/turn/v3"
	"log"
	"net"
	"time"
	"math/rand"
	"github.com/sagernet/sing-box/option"
)

type CommonTurnRelayOptions struct {
    option.ServerOptions
	*option.TurnRelayOptions
}


func genString(serverOptions option.ServerOptions) (string) {
	return fmt.Sprintf("%s:%d", serverOptions.Server, serverOptions.ServerPort)	
}

func ApplyTurnRelay(option CommonTurnRelayOptions)(*Forwarder) {
	if option.TurnRelayOptions == nil {
		return nil
	}
	targetServerAddr := genString(option.ServerOptions) // fmt.Sprintf("%s:%d", option.ServerOptions.Server, option.ServerOptions.ServerPort)
	// Dial TURN Server
	turnServerAddr := genString(option.TurnRelayOptions.ServerOptions) //fmt.Sprintf("%s:%d", option.TurnRelay.Server, option.TurnRelay.Port)
	conn, err := net.Dial("udp", turnServerAddr)
	if err != nil {
		log.Panicf("Failed to connect to TURN server: %s", err)
		return nil
	}

	// Start a new TURN Client and wrap our net.Conn in a STUNConn
	// This allows us to simulate datagram based communication over a net.Conn
	cfg := &turn.ClientConfig{
		STUNServerAddr: turnServerAddr,
		TURNServerAddr: turnServerAddr,
		Conn:           turn.NewSTUNConn(conn),
		Username:       option.TurnRelayOptions.Username,
		Password:       option.TurnRelayOptions.Password,
		Realm:          option.TurnRelayOptions.Realm,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	}

	client, err := turn.NewClient(cfg)
	if err != nil {
		log.Panicf("Failed to create TURN client: %s", err)
		return nil
	}
	defer client.Close()

	// Start listening on the conn provided.
	err = client.Listen()
	if err != nil {
		log.Panicf("Failed to listen: %s", err)
		return nil
	}

	// Allocate a relay socket on the TURN server. On success, it
	// will return a net.PacketConn which represents the remote
	// socket.
	relayConn, err := client.Allocate()
	if err != nil {
		log.Panicf("Failed to allocate: %s", err)
		return nil
	}
	defer func() {
		if closeErr := relayConn.Close(); closeErr != nil {
			log.Fatalf("Failed to close connection: %s", closeErr)
		}
	}()
	rnd_port,err:=getRandomPort(10000,30000)
	if err!=nil{
		log.Panicf("Failed to get random port: %s", err)
		return nil
	}
	// The relayConn's local address is actually the transport
	// address assigned on the TURN server.
	log.Printf("relayed-address=%s", relayConn.LocalAddr().String())
	
	// Forward(src, dst). It's asynchronous.
	
	forwarder, err := Forward(fmt.Sprintf("127.0.0.1:%d",rnd_port), targetServerAddr, relayConn, DefaultTimeout)
	if err != nil {
		return nil
	}
	forwarder.Run()
	option.Server="127.0.0.1"
	option.ServerPort=rnd_port
	return forwarder
}


func getRandomPort(startPort uint16, endPort uint16) (uint16, error) {
    rand.Seed(time.Now().UnixNano())
    for i := 0; i < 100; i++ {
        port := uint16(rand.Intn(int(endPort-startPort+1))) + startPort
        if portAvailable(port) {
            return port, nil
        }
    }
    return 0, fmt.Errorf("Failed to find an available random port in the range %d-%d", startPort, endPort)
}

func portAvailable(port uint16) bool {
    addr := fmt.Sprintf("127.0.0.1:%d", port)
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return false
    }
    listener.Close()
    return true
}
