package store

import (
	log "github.com/sirupsen/logrus"
)

type Manager struct {
	DeskNum   uint      // Index of managed desktop
	ScreenNum uint      // Index of managed screen
	Clients   []*Client // List of clients
}

func CreateManager(deskNum uint, screenNum uint) *Manager {
	return &Manager{
		DeskNum:   deskNum,
		ScreenNum: screenNum,
		Clients:   make([]*Client, 0),
	}
}

func (mg *Manager) AddClient(c *Client) {
	if mg.Exists(c) {
		return
	}

	log.Debug("Add client for manager [", c.Latest.Class, ", workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	mg.Clients = append([]*Client{c}, mg.Clients...)
}

func (mg *Manager) RemoveClient(c *Client) {
	log.Debug("Remove client from manager [", c.Latest.Class, ", workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	idx := mg.Index(mg.Clients, c)
	mg.Clients = append(mg.Clients[:idx], mg.Clients[idx+1:]...)
}

func (mg *Manager) Index(windows []*Client, c *Client) int {

	// Traverse client list
	for i, m := range windows {
		if m.Win.Id == c.Win.Id {
			return i
		}
	}

	return -1
}

func (mg *Manager) Exists(c *Client) bool {
	return mg.Index(mg.Clients, c) >= 0
}
