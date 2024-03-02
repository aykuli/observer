package config

import (
	"flag"
	"log"
	"os"
)

var addr = ServerAddr{"localhost", "8080"}

var filepath = TempFile{Options.FileStoragePath}

type ServerAddr struct {
	Host string
	Port string
}

type TempFile struct {
	path string
}

func parseFlags() {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.Var(&addr, "a", "server address to run on")
	fs.Var(&filepath, "f", "path to save metrics values")
	fs.IntVar(&Options.StoreInterval, "i", 300, "metrics store interval in seconds")
	fs.BoolVar(&Options.Restore, "r", true, "restore metrics from file")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Print(err)
	}

	Options.Address = addr.Host + ":" + addr.Port
	Options.FileStoragePath = filepath.path
	if Options.FileStoragePath == "" {
		Options.SaveMetrics = false
	}
}

func (addr *ServerAddr) String() string {
	return addr.Host + ":" + addr.Port
}

func (addr *ServerAddr) Set(s string) error {
	parsedAddr, err := parseServerAddr(s)
	if err != nil {
		return err
	}

	addr.Host = parsedAddr.Host
	addr.Port = parsedAddr.Port

	return nil
}

func (f *TempFile) String() string {
	return f.path
}

func (f *TempFile) Set(filename string) error {
	_, err := needWriteFile(filename)
	if err != nil {
		return err
	}

	f.path = filename
	return nil
}
