//karing
//go:build windows

package tun

import (
	"strconv"
	"strings"

	"github.com/sagernet/sing-box/log"
	tun "github.com/sagernet/sing-tun"
)

func SetTunnelType(name string) {
	tun.TunnelType = name
}

func newTunWithFallback(logger log.ContextLogger, options *tun.Options, allowInterfaceRename bool) (tun.Tun, error) {
	tunInterface, err := tun.New(*options)
	if err == nil || !allowInterfaceRename || !isWindowsAdapterConflict(err) {
		return tunInterface, err
	}

	baseName := options.Name
	for attempt := 0; attempt < 3; attempt++ {
		fallbackOptions := *options
		fallbackOptions.Name = nextWindowsFallbackInterfaceName(baseName, attempt)
		if logger != nil {
			logger.Warn("retrying Windows TUN adapter creation with fallback interface name: ", fallbackOptions.Name)
		}
		tunInterface, err = tun.New(fallbackOptions)
		if err == nil {
			options.Name = fallbackOptions.Name
			return tunInterface, nil
		}
		if !isWindowsAdapterConflict(err) {
			return nil, err
		}
	}

	return nil, err
}

func isWindowsAdapterConflict(err error) bool {
	if err == nil {
		return false
	}
	errorMessage := err.Error()
	if !strings.Contains(errorMessage, "wintun: Failed to setup adapter") {
		return false
	}
	return strings.Contains(errorMessage, "Cannot create a file when that file already exists") ||
		strings.Contains(errorMessage, "0x000000B7") ||
		strings.Contains(errorMessage, "0xC0000035")
}

func nextWindowsFallbackInterfaceName(baseName string, attempt int) string {
	if baseName == "" {
		baseName = "tun"
	}
	namePrefix := strings.TrimRight(baseName, "0123456789")
	if namePrefix == "" {
		namePrefix = baseName
	}
	nameSuffix := strings.TrimPrefix(baseName, namePrefix)
	baseIndex, _ := strconv.Atoi(nameSuffix)
	return namePrefix + strconv.Itoa(baseIndex+attempt+1)
}
