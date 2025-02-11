package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	runtimeDebug "runtime/debug"
	"sort"
	"strings"
	"syscall"
	"time"

	box "github.com/sagernet/sing-box"
	D "github.com/sagernet/sing-box/common/debug"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/experimental/libbox"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badjson"

	"github.com/spf13/cobra"
)

var commandRun = &cobra.Command{
	Use:   "run",
	Short: "Run service",
	Run: func(cmd *cobra.Command, args []string) {
		err := run()
		if err != nil {
			log.Error(err)  //karing
		}
	},
}

func init() {
	mainCommand.AddCommand(commandRun)
}

type OptionsEntry struct {
	content []byte
	path    string
	options option.Options
}

func readConfigAt(path string) (*OptionsEntry, error) {
	var (
		configContent []byte
		err           error
	)
	if path == "stdin" {
		configContent, err = io.ReadAll(os.Stdin)
	} else {
		configContent, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, E.Cause(err, "read config at ", path)
	}
	options, err := json.UnmarshalExtendedContext[option.Options](globalCtx, configContent)
	if err != nil {
		return nil, E.Cause(err, "decode config at ", path)
	}
	return &OptionsEntry{
		content: configContent,
		path:    path,
		options: options,
	}, nil
}

func readConfig() ([]*OptionsEntry, error) {
	var optionsList []*OptionsEntry
	for _, path := range configPaths {
		optionsEntry, err := readConfigAt(path)
		if err != nil {
			return nil, err
		}
		optionsList = append(optionsList, optionsEntry)
	}
	for _, directory := range configDirectories {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, E.Cause(err, "read config directory at ", directory)
		}
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".json") || entry.IsDir() {
				continue
			}
			optionsEntry, err := readConfigAt(filepath.Join(directory, entry.Name()))
			if err != nil {
				return nil, err
			}
			optionsList = append(optionsList, optionsEntry)
		}
	}
	sort.Slice(optionsList, func(i, j int) bool {
		return optionsList[i].path < optionsList[j].path
	})
	return optionsList, nil
}

func readConfigAndMerge() (option.Options, error) {
	optionsList, err := readConfig()
	if err != nil {
		return option.Options{}, err
	}
	if len(optionsList) == 1 {
		return optionsList[0].options, nil
	}
	var mergedMessage json.RawMessage
	for _, options := range optionsList {
		mergedMessage, err = badjson.MergeJSON(globalCtx, options.options.RawMessage, mergedMessage, false)
		if err != nil {
			return option.Options{}, E.Cause(err, "merge config at ", options.path)
		}
	}
	var mergedOptions option.Options
	err = mergedOptions.UnmarshalJSONContext(globalCtx, mergedMessage)
	if err != nil {
		return option.Options{}, E.Cause(err, "unmarshal merged config")
	}
	return mergedOptions, nil
}

func create() (instance *box.Box,cf context.CancelFunc, err error) { //karing
	defer func() { //karing
		if e := recover(); e != nil {
			content := fmt.Sprintf("%v\n%s", e, string(debug.Stack()))
			err = E.Cause(E.New(content), "panic: create service")
			libbox.SentryCaptureException(&libbox.SentryPanicError{Err: err.Error()})
		}
	}()
	stacks := D.Stacks(false, false) //karing
	if len(stacks) > 0 {  //karing
		for key := range stacks {
			D.MainGoId = key
			break
		}
	}
	
	options, err := readConfigAndMerge()
	if err != nil {
		libbox.SentryCaptureException(err) //karing
		return nil, nil, err
	}
	if disableColor {
		if options.Log == nil {
			options.Log = &option.LogOptions{}
		}
		options.Log.DisableColor = true
	}
	ctx, cancel := context.WithCancel(globalCtx)
	instance, err = box.New(box.Options{ //karing
		Context: ctx,
		Options: options,
	})
	if err != nil {
		cancel()
		libbox.SentryCaptureException(E.Cause(err, "create service")) //karing
		return nil, nil, E.Cause(err, "create service")
	}

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer func() {
		signal.Stop(osSignals)
		close(osSignals)
	}()
	startCtx, finishStart := context.WithCancel(context.Background())
	go func() {
		_, loaded := <-osSignals
		if loaded {
			cancel()
			closeMonitor(startCtx)
		}
	}()
	err = instance.Start()
	finishStart()
	if err != nil {
		cancel()
		libbox.SentryCaptureException(E.Cause(err, "start service")) //karing
		return nil, nil, E.Cause(err, "start service")
	}
	if servicePort != 0 { //karing
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", servicePort)) 
		if err == nil{
			conn.Close()
		}
	}

	return instance, cancel, nil
}

func run() error {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(osSignals)
	for {
		instance, cancel, err := create()
		if err != nil {
			return err
		}
		runtimeDebug.FreeOSMemory()
		for {
			select{  //karing
			case osSignal := <-osSignals:
				if osSignal == syscall.SIGHUP {
					err = check()
					if err != nil {
						log.Error(E.Cause(err, "reload service"))
						continue
					}
				}
				cancel()
				closeCtx, closed := context.WithCancel(context.Background())
				go closeMonitor(closeCtx)
				err = instance.Close()
				closed()
				if osSignal != syscall.SIGHUP {
					if err != nil {
						log.Error(E.Cause(err, "sing-box did not closed properly"))
					}
					return nil
				}
				break
			case <-instance.Quit:  //karing
				cancel()
				closeCtx, closed := context.WithCancel(context.Background())
				go closeMonitor(closeCtx)
				go func() {
					time.Sleep(3 * time.Second)
					os.Exit(1)
				}()
				instance.Close()
				closed()
				return nil
			}
		}
	}
}

func closeMonitor(ctx context.Context) {
	time.Sleep(C.FatalStopTimeout)
	select {
	case <-ctx.Done():
		return
	default:
	}
	log.Fatal("sing-box did not close!")
}