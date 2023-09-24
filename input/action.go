package input

import (
	"os"
	"os/exec"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/seyys/sticky-display/common"
	"github.com/seyys/sticky-display/desktop"

	log "github.com/sirupsen/logrus"
)

var (
	executeCallbacksFun []func(string) // Execute events callback functions
)

func Execute(action string, mod string, tr *desktop.Tracker) bool {
	success := false
	if len(strings.TrimSpace(action)) == 0 {
		return false
	}

	log.Info("Execute action [", action, "-", mod, "]")

	for _, ws := range tr.Workspaces {
		active := tr.ActiveWorkspace()

		// Execute only on active screen
		if mod == "current" && ws.Location != active.Location {
			continue
		}

		// Choose action command
		switch action {
		case "enable":
			success = Enable(tr, ws)
		case "exit":
			success = Exit(tr)
		default:
			success = External(action)
		}

		if !success {
			return false
		}

		// Notify socket
		type Action struct {
			Desk   uint
			Screen uint
		}
		NotifySocket(Message[Action]{
			Type: "Action",
			Name: action,
			Data: Action{Screen: ws.Location.ScreenNum},
		})
	}

	// Execute callbacks
	executeCallbacks(action)

	return true
}

func Query(state string, tr *desktop.Tracker) bool {
	success := false
	if len(strings.TrimSpace(state)) == 0 {
		return false
	}

	log.Info("Query state [", state, "]")

	ws := tr.ActiveWorkspace()

	// Choose state query
	switch state {
	case "workspaces":
		type Workspaces struct {
			Desk       uint
			Screen     uint
			Workspaces []*desktop.Workspace
		}
		NotifySocket(Message[Workspaces]{
			Type: "State",
			Name: state,
			Data: Workspaces{Screen: ws.Location.ScreenNum, Workspaces: maps.Values(tr.Workspaces)},
		})
		success = true
	case "arguments":
		NotifySocket(Message[common.Arguments]{
			Type: "State",
			Name: state,
			Data: common.Args,
		})
		success = true
	case "configs":
		NotifySocket(Message[common.Configuration]{
			Type: "State",
			Name: state,
			Data: common.Config,
		})
		success = true
	}

	return success
}

func Enable(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	tr.Update()

	return true
}

func Exit(tr *desktop.Tracker) bool {
	log.Info("Exit")

	os.Remove(common.Args.Sock + ".in")
	os.Remove(common.Args.Sock + ".out")

	os.Exit(1)

	return true
}

func External(command string) bool {
	params := strings.Split(command, " ")

	log.Info("Execute command ", params[0], " ", params[1:])

	// Execute external command
	cmd := exec.Command(params[0], params[1:]...)
	if err := cmd.Run(); err != nil {
		log.Error(err)
		return false
	}

	return true
}

func OnExecute(fun func(string)) {
	executeCallbacksFun = append(executeCallbacksFun, fun)
}

func executeCallbacks(arg string) {
	log.Info("Execute event ", arg)

	for _, fun := range executeCallbacksFun {
		fun(arg)
	}
}
