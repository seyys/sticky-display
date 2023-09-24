package desktop

import "github.com/seyys/sticky-display/store"

type Layout interface {
	AddClient(c *store.Client)
	RemoveClient(c *store.Client)
	GetManager() *store.Manager
}
