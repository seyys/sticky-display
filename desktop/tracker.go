package desktop

import (
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/seyys/sticky-display/common"
	"github.com/seyys/sticky-display/store"

	log "github.com/sirupsen/logrus"
)

type Tracker struct {
	Clients    map[xproto.Window]*store.Client // List of clients that are being tracked
	Workspaces map[Location]*Workspace         // List of workspaces per location
	Action     chan string                     // Event channel for actions
	Handler    *Handler                        // Helper for event handlers
}

type Location struct {
	ScreenNum uint // Workspace screen number
}

type Handler struct {
	Timer      *time.Timer    // Timer to handle delayed structure events
	SwapScreen *HandlerClient // Stores client for screen swap
}

type HandlerClient struct {
	Active bool          // Indicates active handler event
	Source *store.Client // Stores moving/resizing client
	Target *store.Client // Stores hovered client
}

func CreateTracker(ws map[Location]*Workspace) *Tracker {
	tr := Tracker{
		Clients:    make(map[xproto.Window]*store.Client),
		Workspaces: ws,
		Action:     make(chan string),
		Handler: &Handler{
			SwapScreen: &HandlerClient{},
		},
	}

	// Attach to root events
	store.OnStateUpdate(tr.onStateUpdate)
	store.OnPointerUpdate(tr.onPointerUpdate)

	// Update on startup
	tr.Update()

	return &tr
}

func (tr *Tracker) Update() {
	log.Debug("Update trackable clients [", len(tr.Clients), "/", len(store.Windows), "]")

	// Map trackable windows
	trackable := make(map[xproto.Window]bool)
	for _, w := range store.Windows {
		trackable[w] = tr.isTrackable(w)
	}

	// Remove untrackable windows
	for w := range tr.Clients {
		if !trackable[w] {
			tr.untrackWindow(w)
		}
	}

	// Add trackable windows
	for _, w := range store.Windows {
		if trackable[w] {
			tr.trackWindow(w)
		}
	}
}

func (tr *Tracker) Reset() {
	log.Debug("Reset trackable clients [", len(tr.Clients), "/", len(store.Windows), "]")

	// Reset client list
	for w := range tr.Clients {
		tr.untrackWindow(w)
	}

	// Reset workspaces
	tr.Workspaces = CreateWorkspaces()
}

func (tr *Tracker) ActiveWorkspace() *Workspace {
	location := Location{ScreenNum: store.CurrentScreen}

	// Validate active workspace
	ws := tr.Workspaces[location]
	if ws == nil {
		log.Warn("Invalid active screen [", location.ScreenNum, "]")
	}

	return ws
}

func (tr *Tracker) ClientWorkspace(c *store.Client) *Workspace {
	location := Location{ScreenNum: c.Latest.ScreenNum}

	// Validate client workspace
	ws := tr.Workspaces[location]
	if ws == nil {
		log.Warn("Invalid client screen [", location.ScreenNum, "]")
	}

	return ws
}

func (tr *Tracker) trackWindow(w xproto.Window) bool {
	if tr.isTracked(w) {
		return false
	}

	// Add new client
	c := store.CreateClient(w)
	tr.Clients[c.Win.Id] = c
	ws := tr.ClientWorkspace(c)
	ws.AddClient(c)
	c.Pin()

	// Attach handlers
	tr.attachHandlers(c)

	return true
}

func (tr *Tracker) untrackWindow(w xproto.Window) bool {
	if !tr.isTracked(w) {
		return false
	}

	// Client and workspace
	c := tr.Clients[w]
	ws := tr.ClientWorkspace(c)

	// Detach events
	xevent.Detach(store.X, w)

	// Restore client
	c.Restore(false)

	// Remove client
	ws.RemoveClient(c)
	delete(tr.Clients, w)
	c.Pin()

	return true
}

func (tr *Tracker) handleMoveClient(c *store.Client) {
	// ws := tr.ClientWorkspace(c)
	if !tr.isTracked(c.Win.Id) || store.IsMaximized(c.Win.Id) {
		return
	}

	// Previous position
	pGeom := c.Latest.Dimensions.Geometry
	px, py, pw, ph := pGeom.Pieces()

	// Current position
	cGeom, err := c.Win.DecorGeometry()
	if err != nil {
		return
	}
	cx, cy, cw, ch := cGeom.Pieces()

	// Check position change
	moved := cx != px || cy != py
	resized := cw != pw || ch != ph
	active := c.Win.Id == store.ActiveWindow

	if active && moved && !resized {
		log.Debug("Client move handler fired [", c.Latest.Class, "]")

		// Check if pointer moves to another screen
		tr.Handler.SwapScreen.Active = false
		if c.Latest.ScreenNum != store.CurrentScreen {
			tr.Handler.SwapScreen = &HandlerClient{Active: true, Source: c}
		}
	}
}

func (tr *Tracker) handleWorkspaceChange(c *store.Client) {
	if !tr.isTracked(c.Win.Id) {
		return
	}
	log.Debug("Client workspace handler fired [", c.Latest.Class, "]")

	// Remove client from current workspace
	ws := tr.ClientWorkspace(c)
	ws.RemoveClient(c)

	// Reset screen swapping event
	tr.Handler.SwapScreen.Active = false

	// Update client desktop and screen
	if !tr.isTrackable(c.Win.Id) {
		return
	}
	c.Update()
	if common.IsInIntList(int(c.Latest.ScreenNum), common.Config.StickyDisplays) {
		c.Pin()
	} else {
		c.UnPin()
	}

	// Add client to new workspace
	ws = tr.ClientWorkspace(c)
	ws.AddClient(c)
	c.Restore(false)
}

func (tr *Tracker) onStateUpdate(aname string) {
	viewportChanged := common.IsInList(aname, []string{"_NET_NUMBER_OF_DESKTOPS", "_NET_DESKTOP_LAYOUT", "_NET_DESKTOP_GEOMETRY", "_NET_DESKTOP_VIEWPORT", "_NET_WORKAREA"})
	clientsChanged := common.IsInList(aname, []string{"_NET_CLIENT_LIST_STACKING", "_NET_ACTIVE_WINDOW"})

	// Viewport changed or clients changed
	if viewportChanged || clientsChanged {
		tr.Update()

		// Deactivate handlers
		tr.Handler.SwapScreen.Active = false
	}
}

func (tr *Tracker) onPointerUpdate(button uint16) {
	// Reset timer
	if tr.Handler.Timer != nil {
		tr.Handler.Timer.Stop()
	}

	// Wait on button release
	var t time.Duration = 0
	if button == 0 {
		t = 50
	}

	// Wait for structure events
	tr.Handler.Timer = time.AfterFunc(t*time.Millisecond, func() {

		// Window moved to another screen
		if tr.Handler.SwapScreen.Active {
			tr.handleWorkspaceChange(tr.Handler.SwapScreen.Source)
		}
	})
}

func (tr *Tracker) attachHandlers(c *store.Client) {
	c.Win.Listen(xproto.EventMaskStructureNotify | xproto.EventMaskPropertyChange | xproto.EventMaskFocusChange)

	// Attach structure events
	xevent.ConfigureNotifyFun(func(x *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
		log.Trace("Client structure event [", c.Latest.Class, "]")

		// Handle structure events
		tr.handleMoveClient(c)
	}).Connect(store.X, c.Win.Id)

	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		aname, _ := xprop.AtomName(store.X, ev.Atom)
		log.Trace("Client property event ", aname, " [", c.Latest.Class, "]")
		// TODO prevent unsetting sticky in selected display
	}).Connect(store.X, c.Win.Id)
}

func (tr *Tracker) isTracked(w xproto.Window) bool {
	_, ok := tr.Clients[w]
	return ok
}

func (tr *Tracker) isTrackable(w xproto.Window) bool {
	info := store.GetInfo(w)
	return !store.IsSpecial(info) && !store.IsIgnored(info)
}
