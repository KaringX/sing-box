//go:build with_karing

package main

import (
	"context"
	"os"
	"os/user"
	"strconv"
	"time"

	"github.com/sagernet/sing-box/experimental/libbox"
	_ "github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/service/filemanager"
	"github.com/valyala/fastjson"

	"github.com/spf13/cobra"
)

var (
	globalCtx         context.Context
	configPaths       []string
	configDirectories []string
	workingDir        string
	disableColor      bool
	serviceConfigPath string  
	servicePort       int
)

var mainCommand = &cobra.Command{
	Use:              "karing",
	PersistentPreRun: preRun,
}

func init() {
	mainCommand.PersistentFlags().StringArrayVarP(&configPaths, "config", "c", nil, "set configuration file path")
	mainCommand.PersistentFlags().StringArrayVarP(&configDirectories, "config-directory", "C", nil, "set configuration directory path")
	mainCommand.PersistentFlags().StringVarP(&workingDir, "directory", "D", "", "set working directory")
	mainCommand.PersistentFlags().BoolVarP(&disableColor, "disable-color", "", false, "disable color output")
	mainCommand.PersistentFlags().StringVarP(&serviceConfigPath, "service-config", "s", "", "service-config")
	mainCommand.PersistentFlags().IntVarP(&servicePort, "service-port", "p", 0, "service-port")
}

func preRun(cmd *cobra.Command, args []string) {
	libbox.SentryInit(serviceConfigPath)
	err := setUpDir()
	if(err != nil){
		log.Fatal(err)
	}
	globalCtx = context.Background()
	sudoUser := os.Getenv("SUDO_USER")
	sudoUID, _ := strconv.Atoi(os.Getenv("SUDO_UID"))
	sudoGID, _ := strconv.Atoi(os.Getenv("SUDO_GID"))
	if sudoUID == 0 && sudoGID == 0 && sudoUser != "" {
		sudoUserObject, _ := user.Lookup(sudoUser)
		if sudoUserObject != nil {
			sudoUID, _ = strconv.Atoi(sudoUserObject.Uid)
			sudoGID, _ = strconv.Atoi(sudoUserObject.Gid)
		}
	}
	if sudoUID > 0 && sudoGID > 0 {
		globalCtx = filemanager.WithDefault(globalCtx, "", "", "", sudoUID, sudoGID)
	}
	if disableColor {
		log.SetStdLogger(log.NewDefaultFactory(context.Background(), log.Formatter{BaseTime: time.Now(), DisableColors: true}, os.Stderr, "", nil, false).Logger())
	}
	if workingDir != "" {
		_, err := os.Stat(workingDir)
		if err != nil {
			filemanager.MkdirAll(globalCtx, workingDir, 0o777)
		}
		err = os.Chdir(workingDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	if len(configPaths) == 0 && len(configDirectories) == 0 {
		configPaths = append(configPaths, "config.json")
	}
}

func setUpDir() error {
	content, err := os.ReadFile(serviceConfigPath)
	if err != nil {
		return err
	}
	var parser fastjson.Parser
	value, err1 := parser.ParseBytes(content)
	if err1 != nil {
		return err1
	}

	base_dir := stringNotNil(value.GetStringBytes("base_dir"))
	work_dir := stringNotNil(value.GetStringBytes("work_dir"))
	cache_dir := stringNotNil(value.GetStringBytes("cache_dir"))
	core_path := stringNotNil(value.GetStringBytes("core_path"))
	if len(core_path) == 0 {
		return E.New(serviceConfigPath + " : core_path is empty")
	}
	libbox.Setup(base_dir, work_dir, cache_dir, false)
	configPaths = append(configPaths, core_path)
	return nil
}
func stringNotNil(v []byte) string {
	if v == nil {
		return ""
	}
	return string(v)
}