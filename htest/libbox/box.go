package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	B "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/common/urltest"
	"github.com/sagernet/sing-box/experimental/libbox"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/filemanager"
	"github.com/sagernet/sing/service/pause"
)

func main() {
	options := option.Options{}
	content, err := os.ReadFile("./htest/a.json")
	json.Unmarshal(content, &options)
	fmt.Println(string(content))
	libbox.Setup("./htest/", "./htest/", "./htest/tmp/", true)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = filemanager.WithDefault(ctx, "./htest/", "./htest/tmp/", os.Getuid(), os.Getgid())
	urlTestHistoryStorage := urltest.NewHistoryStorage()
	ctx = service.ContextWithPtr(ctx, urlTestHistoryStorage)

	instance, err := B.New(B.Options{
		Context: ctx,
		Options: options,
	})
	if err != nil {
		fmt.Println(err)
		cancel()
		return
	}
	commandServer = libbox.NewCommandServer(&CommandServerHandler{}, 100)
	commandServer.Start()
	libservice := libbox.NewBoxService(
		ctx,
		cancel,
		instance,
		service.FromContext[pause.Manager](ctx),
		urlTestHistoryStorage,
	)
	// instance.Start()
	libservice.Start()
	<-time.After(100 * time.Second)
	commandServer.SetService(&libservice)
	<-time.Tick(1 * time.Second)
	fmt.Println("command group update")
	// pm:=service.FromContext[pause.Manager](ctx)

	groupInfoOnlyClient := libbox.NewCommandClient(
		&CommandClientHandler{
			logger: log.NewNOPFactory().NewLogger("[GroupInfoOnly Command Client]"),
		},
		&libbox.CommandClientOptions{
			Command:        libbox.CommandGroupInfoOnly,
			StatusInterval: 300000000, //300ms debounce
		},
	)
	groupInfoOnlyClient.Connect()
	groupClient := libbox.NewCommandClient(
		&CommandClientHandler{
			logger: log.NewNOPFactory().NewLogger("[GroupInfoOnly Command Client]"),
		},
		&libbox.CommandClientOptions{
			Command:        libbox.CommandGroup,
			StatusInterval: 300000000, //300ms debounce
		},
	)
	groupClient.Connect()
	for i := 0; i < 4; i++ {
		<-time.Tick(1000 * time.Millisecond)
		fmt.Println("selecting auto")
		libbox.NewStandaloneCommandClient().SelectOutbound("Select", "Auto")
		<-time.Tick(1000 * time.Millisecond)
		fmt.Println("selecting outbound")
		libbox.NewStandaloneCommandClient().SelectOutbound("Select", "test")
	}
	fmt.Println("===========Finished many bounce")
	<-time.Tick(2000 * time.Millisecond)
	fmt.Println("===========selecting final auto")
	libbox.NewStandaloneCommandClient().SelectOutbound("Select", "Auto")
	fmt.Println("===========Closing")
	instance.Close()
	instance.Close()
	libservice.Close()
	commandServer.Close()
	<-time.After(1 * time.Second)
	options = option.Options{}
	content, err = os.ReadFile("./htest/a.json")
	json.Unmarshal(content, &options)
	fmt.Println(string(content))
	libbox.Setup("./htest/", "./htest/", "./htest/", true)

	urlTestHistoryStorage = urltest.NewHistoryStorage()

	instance, err = B.New(B.Options{
		Context: ctx,
		Options: options,
	})
	if err != nil {
		cancel()
		return
	}
	commandServer = libbox.NewCommandServer(&CommandServerHandler{}, 100)
	commandServer.Start()
	libservice = libbox.NewBoxService(
		ctx,
		cancel,
		instance,
		service.FromContext[pause.Manager](ctx),
		urlTestHistoryStorage,
	)
	// instance.Start()
	libservice.Start()

	commandServer.SetService(&libservice)

	<-time.Tick(100 * time.Second)
	fmt.Println("command group update")

	groupInfoOnlyClient = libbox.NewCommandClient(
		&CommandClientHandler{
			logger: log.NewNOPFactory().NewLogger("[GroupInfoOnly Command Client]"),
		},
		&libbox.CommandClientOptions{
			Command:        libbox.CommandGroupInfoOnly,
			StatusInterval: 3000000000, //300ms debounce
		},
	)
	groupInfoOnlyClient.Connect()

	groupClient.Connect()
	for i := 0; i < 4; i++ {
		<-time.Tick(1000 * time.Millisecond)
		fmt.Println("selecting auto")
		libbox.NewStandaloneCommandClient().SelectOutbound("Select", "Auto")
		<-time.Tick(1000 * time.Millisecond)
		fmt.Println("selecting outbound")
		libbox.NewStandaloneCommandClient().SelectOutbound("Select", "test")
	}
	fmt.Println("===========Finished many bounce")
	<-time.Tick(20000 * time.Millisecond)
	fmt.Println("===========selecting final auto")
	libbox.NewStandaloneCommandClient().SelectOutbound("Select", "Auto")

	libservice.Close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	<-sigCh

}
