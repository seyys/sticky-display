package store

import (
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/seyys/sticky-display/common"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	Win      *xwindow.Window `json:"-"` // X window object
	Created  time.Time       // Internal client creation time
	Original *Info           // Original client window information
	Latest   *Info           // Latest client window information
}

type Info struct {
	Class      string     // Client window application name
	Name       string     // Client window title name
	DeskNum    uint       // Client window desktop
	ScreenNum  uint       // Client window screen
	Types      []string   // Client window types
	States     []string   // Client window states
	Dimensions Dimensions // Client window dimensions
}

type Dimensions struct {
	Geometry xrect.Rect        // Client window geometry
	Hints    Hints             // Client window dimension hints
	Extents  ewmh.FrameExtents // Client window geometry extents
	AdjPos   bool              // Adjust position on move/resize
	AdjSize  bool              // Adjust size on move/resize
}

type Hints struct {
	Normal icccm.NormalHints // Client window geometry hints
	Motif  motif.Hints       // Client window decoration hints
}

func CreateClient(w xproto.Window) *Client {
	i := GetInfo(w)
	c := &Client{
		Win:      xwindow.New(X, w),
		Created:  time.Now(),
		Original: i,
		Latest:   i,
	}

	// Restore window decorations
	c.Restore(false)

	return c
}

func (c *Client) Activate() {
	ewmh.ActiveWindowReq(X, c.Win.Id)
}

func (c *Client) Pin() {
	if common.IsInIntList(int(c.Latest.ScreenNum), common.Config.StickyDisplays) {
		ewmh.WmStateReq(X, c.Win.Id, 1, "_NET_WM_STATE_STICKY")
	}
}

func (c *Client) UnPin() {
	// TODO restore original sticky state
	ewmh.WmStateReq(X, c.Win.Id, 0, "_NET_WM_STATE_STICKY")
}

func (c *Client) MoveResize(x, y, w, h int) {

	// Calculate dimension offsets
	ext := c.Latest.Dimensions.Extents
	dx, dy, dw, dh := 0, 0, 0, 0

	if c.Latest.Dimensions.AdjPos {
		dx, dy = ext.Left, ext.Top
	}
	if c.Latest.Dimensions.AdjSize {
		dw, dh = ext.Left+ext.Right, ext.Top+ext.Bottom
	}

	// Move and resize window
	err := ewmh.MoveresizeWindow(X, c.Win.Id, x+dx, y+dy, w-dw, h-dh)
	if err != nil {
		log.Warn("Error on window move/resize [", c.Latest.Class, "]")
	}

	// Update stored dimensions
	c.Update()
}

func (c *Client) Update() {
	info := GetInfo(c.Win.Id)
	if len(info.Class) == 0 {
		return
	}

	// Update client info
	log.Debug("Update client info [", info.Class, "]")
	c.Latest = info
}

func (c *Client) Restore(original bool) {
	c.Update()
}

func IsSpecial(info *Info) bool {

	// Check internal windows
	if info.Class == common.Build.Name {
		log.Info("Ignore internal window [", info.Class, "]")
		return true
	}

	// Check window types
	types := []string{
		"_NET_WM_WINDOW_TYPE_DOCK",
		"_NET_WM_WINDOW_TYPE_DESKTOP",
		"_NET_WM_WINDOW_TYPE_TOOLBAR",
		"_NET_WM_WINDOW_TYPE_UTILITY",
		"_NET_WM_WINDOW_TYPE_TOOLTIP",
		"_NET_WM_WINDOW_TYPE_SPLASH",
		"_NET_WM_WINDOW_TYPE_DIALOG",
		"_NET_WM_WINDOW_TYPE_COMBO",
		"_NET_WM_WINDOW_TYPE_NOTIFICATION",
		"_NET_WM_WINDOW_TYPE_DROPDOWN_MENU",
		"_NET_WM_WINDOW_TYPE_POPUP_MENU",
		"_NET_WM_WINDOW_TYPE_MENU",
		"_NET_WM_WINDOW_TYPE_DND",
	}
	for _, typ := range info.Types {
		if common.IsInList(typ, types) {
			log.Info("Ignore window with type ", typ, " [", info.Class, "]")
			return true
		}
	}

	return false
}

func IsIgnored(info *Info) bool {

	// Check ignored windows
	for _, s := range common.Config.WindowIgnore {
		conf_class := s[0]
		conf_name := s[1]

		reg_class := regexp.MustCompile(strings.ToLower(conf_class))
		reg_name := regexp.MustCompile(strings.ToLower(conf_name))

		// Ignore all windows with this class
		class_match := reg_class.MatchString(strings.ToLower(info.Class))

		// But allow the window with a special name
		name_match := conf_name != "" && reg_name.MatchString(strings.ToLower(info.Name))

		if class_match && !name_match {
			log.Info("Ignore window with ", strings.TrimSpace(strings.Join(s, " ")), " from config [", info.Class, "]")
			return true
		}
	}

	return false
}

func IsMaximized(w xproto.Window) bool {
	info := GetInfo(w)

	// Check maximized windows
	for _, state := range info.States {
		if strings.HasPrefix(state, "_NET_WM_STATE_MAXIMIZED") {
			log.Info("Ignore maximized window [", info.Class, "]")
			return true
		}
	}

	return false
}

func GetInfo(w xproto.Window) *Info {
	var err error

	var class string
	var name string
	var deskNum uint
	var screenNum uint
	var types []string
	var states []string
	var dimensions Dimensions

	// Window class (internal class name of the window)
	cls, err := icccm.WmClassGet(X, w)
	if err != nil {
		log.Trace("Error on request ", err)
	} else if cls != nil {
		class = cls.Class
	}

	// Window name (title on top of the window)
	name, err = icccm.WmNameGet(X, w)
	if err != nil {
		name = class
	}

	screenNum = GetScreenNum(w)

	// Window types (types of the window)
	types, err = ewmh.WmWindowTypeGet(X, w)
	if err != nil {
		types = []string{}
	}

	// Window states (states of the window)
	states, err = ewmh.WmStateGet(X, w)
	if err != nil {
		states = []string{}
	}

	// Window geometry (dimensions of the window)
	geometry, err := xwindow.New(X, w).DecorGeometry()
	if err != nil {
		geometry = &xrect.XRect{}
	}

	// Window normal hints (normal hints of the window)
	nhints, err := icccm.WmNormalHintsGet(X, w)
	if err != nil {
		nhints = &icccm.NormalHints{}
	}

	// Window motif hints (hints of the window)
	mhints, err := motif.WmHintsGet(X, w)
	if err != nil {
		mhints = &motif.Hints{}
	}

	// Window extents (server/client decorations of the window)
	extNet, _ := xprop.PropValNums(xprop.GetProperty(X, w, "_NET_FRAME_EXTENTS"))
	extGtk, _ := xprop.PropValNums(xprop.GetProperty(X, w, "_GTK_FRAME_EXTENTS"))

	ext := make([]uint, 4)
	for i, e := range extNet {
		ext[i] += e
	}
	for i, e := range extGtk {
		ext[i] -= e
	}

	// Window dimensions (geometry/extent information for move/resize)
	dimensions = Dimensions{
		Geometry: geometry,
		Hints: Hints{
			Normal: *nhints,
			Motif:  *mhints,
		},
		Extents: ewmh.FrameExtents{
			Left:   int(ext[0]),
			Right:  int(ext[1]),
			Top:    int(ext[2]),
			Bottom: int(ext[3]),
		},
		AdjPos:  (extNet != nil && mhints.Flags&motif.HintDecorations > 0 && mhints.Decoration > 1) || (extGtk != nil),
		AdjSize: (extNet != nil) || (extGtk != nil),
	}

	return &Info{
		Class:      class,
		Name:       name,
		DeskNum:    deskNum,
		ScreenNum:  screenNum,
		Types:      types,
		States:     states,
		Dimensions: dimensions,
	}
}

func GetScreenNum(w xproto.Window) uint {

	// Outer window dimensions
	geom, err := xwindow.New(X, w).DecorGeometry()
	if err != nil {
		return 0
	}

	// Window center position
	center := &common.Pointer{
		X: int16(geom.X() + (geom.Width() / 2)),
		Y: int16(geom.Y() + (geom.Height() / 2)),
	}

	return ScreenNumGet(center)
}
