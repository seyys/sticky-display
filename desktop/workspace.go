package desktop

import (
	"github.com/seyys/sticky-display/store"

	log "github.com/sirupsen/logrus"
)

type Workspace struct {
	Location        Location // Desktop and screen location
	Layouts         []Layout // List of available layouts
	ActiveLayoutNum uint     // Active layout index
}

func CreateWorkspaces() map[Location]*Workspace {
	workspaces := make(map[Location]*Workspace)

	for screenNum := uint(0); screenNum < store.ScreenCount; screenNum++ {
		location := Location{ScreenNum: screenNum}

		// Create layouts for each desktop and screen
		ws := &Workspace{
			Location:        location,
			ActiveLayoutNum: 0,
		}

		// Map location to workspace
		workspaces[location] = ws
	}

	return workspaces
}

func (ws *Workspace) ActiveLayout() Layout {
	return ws.Layouts[ws.ActiveLayoutNum]
}

func (ws *Workspace) Restore(original bool) {
	mg := ws.ActiveLayout().GetManager()

	log.Info("Untile ", len(mg.Clients), " windows [workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	// Restore client dimensions
	for _, c := range mg.Clients {
		c.Restore(original)
	}
}

func (ws *Workspace) AddClient(c *store.Client) {
	log.Info("Add client for each layout [", c.Latest.Class, "]")

	// Add client to all layouts
	for _, l := range ws.Layouts {
		l.AddClient(c)
	}
}

func (ws *Workspace) RemoveClient(c *store.Client) {
	log.Info("Remove client from each layout [", c.Latest.Class, "]")

	// Remove client from all layouts
	for _, l := range ws.Layouts {
		l.RemoveClient(c)
	}
}
