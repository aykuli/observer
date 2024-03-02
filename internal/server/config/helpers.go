package config

import (
	"errors"
	"strings"
)

func parseServerAddr(s string) (ServerAddr, error) {
	hp := strings.Split(s, ":")

	if len(hp) != 2 {
		return ServerAddr{}, errors.New("need address in a form host:port")
	}

	if hp[0] == "" {
		hp[0] = hostDefault
	}

	if hp[1] == "" {
		hp[1] = portDefault
	}

	return ServerAddr{
		Host: hp[0],
		Port: hp[1],
	}, nil
}

func needWriteFile(filename string) (bool, error) {
	if filename == "" {
		return false, nil
	}

	return true, nil
}
