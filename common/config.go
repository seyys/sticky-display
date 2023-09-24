package common

import (
	"fmt"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"

	log "github.com/sirupsen/logrus"
)

var (
	Config Configuration // Decoded config values
)

type Configuration struct {
	StickyDisplays []int             `toml:"sticky_displays"` // Display numbers to sticky windows on
	WindowIgnore   [][]string        `toml:"window_ignore"`   // Regex to ignore windows
	Keys           map[string]string `toml:"keys"`            // Event bindings for keyboard shortcuts
}

func InitConfig() {

	// Create config folder if not exists
	configFolderPath := filepath.Dir(Args.Config)
	if _, err := os.Stat(configFolderPath); os.IsNotExist(err) {
		os.MkdirAll(configFolderPath, 0700)
	}

	// Write default config if not exists
	if _, err := os.Stat(Args.Config); os.IsNotExist(err) {
		ioutil.WriteFile(Args.Config, File.Toml, 0644)
	}

	// Read config file into memory
	readConfig(Args.Config)

	// Config file watcher
	watchConfig(Args.Config)
}

func ConfigFilePath(name string) string {

	// Obtain user home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error obtaining home directory ", err)
	}
	configFolderPath := filepath.Join(userHome, ".config", name)

	// Obtain config directory
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		configFolderPath = filepath.Join(xdgConfigHome, name)
	}

	return filepath.Join(configFolderPath, "config.toml")
}

func readConfig(configFilePath string) {
	fmt.Println(fmt.Errorf("LOAD %s [%s]", configFilePath, Build.Summary))
	log.Info("Starting [", Build.Summary, "]")

	// Decode contents into struct
	toml.DecodeFile(configFilePath, &Config)
}

func watchConfig(configFilePath string) {

	// Init file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	} else {
		watcher.Add(configFilePath)
	}

	// Listen for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					readConfig(configFilePath)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error(err)
			}
		}
	}()
}
