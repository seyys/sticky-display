package main

import (
	_ "embed"

	"fmt"
	"io"
	"os"
	"runtime/debug"
	"syscall"

	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/seyys/sticky-display/common"
	"github.com/seyys/sticky-display/desktop"
	"github.com/seyys/sticky-display/input"
	"github.com/seyys/sticky-display/store"

	log "github.com/sirupsen/logrus"
)

var (
	// Build name
	name = "sticky-display"

	// Build version
	version = "0.1"

	// Build commit
	commit = "local"

	// Build date
	date = "unknown"
)

var (
	//go:embed config.toml
	toml []byte
)

func main() {

	// Init command line arguments
	common.InitArgs(name, version, commit, date)

	// Init embedded files
	common.InitFiles(toml)

	// Init lock and log files
	defer InitLock().Close()
	InitLog()

	// Init config and root
	common.InitConfig()
	store.InitRoot()

	prepare()

	if common.Args.PrintDisplay {
		println("This is display", store.CurrentScreen)
		return
	}

	// Run X event loop
	xevent.Main(store.X)
}

func prepare() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(fmt.Errorf("%s\n%s", err, debug.Stack()))
		}
	}()

	// Create workspaces and tracker
	workspaces := desktop.CreateWorkspaces()
	tracker := desktop.CreateTracker(workspaces)

	// Bind input events
	input.BindSignal(tracker)
	input.BindSocket(tracker)
	input.BindMouse(tracker)
	input.BindKeys(tracker)
}

func InitLock() *os.File {
	file, err := createLockFile(common.Args.Lock)
	if err != nil {
		fmt.Println(fmt.Errorf("%s already running (%s)", common.Build.Name, err))
		os.Exit(1)
	}

	return file
}

func InitLog() *os.File {
	if common.Args.VVV {
		log.SetLevel(log.TraceLevel)
	} else if common.Args.VV {
		log.SetLevel(log.DebugLevel)
	} else if common.Args.V {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})

	file, err := createLogFile(common.Args.Log)
	if err != nil {
		return file
	}

	log.SetOutput(io.MultiWriter(os.Stderr, file))
	log.RegisterExitHandler(func() {
		if file != nil {
			file.Close()
		}
	})

	return file
}

func createLockFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(fmt.Errorf("FILE error (%s)", err))
		return nil, nil
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}

func createLogFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Errorf("FILE error (%s)", err))
		return nil, err
	}

	return file, nil
}
