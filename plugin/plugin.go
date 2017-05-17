package plugin

import (
	"plugin"

	"github.com/themotion/ladder/config"
	"github.com/themotion/ladder/log"
)

// Loader is an interface that needs to be satisfied to load plugins
type Loader interface {
	LoadFromConfig(cfg *config.Config) error
	Load(path string) (*plugin.Plugin, error)
}

// BaseLoader will load Ladder plugins
type BaseLoader struct {
	// Plugins are the loaded plugins by the loader
	Plugins map[string]*plugin.Plugin
	log     *log.Log // custom logger
}

// NewBaseLoader returns a new plugin loader
func NewBaseLoader() (*BaseLoader, error) {
	return &BaseLoader{
		Plugins: map[string]*plugin.Plugin{},
		log:     log.Logger,
	}, nil
}

// LoadFromConfig will load all the plugins from a config file
func (l *BaseLoader) LoadFromConfig(cfg *config.Config) error {
	if len(cfg.Global.Plugins) < 1 {
		return nil
	}

	for _, pp := range cfg.Global.Plugins {
		p, err := l.Load(pp)
		if err != nil {
			return err
		}
		l.Plugins[pp] = p
	}

	return nil
}

// Load will load a plugin
func (l *BaseLoader) Load(path string) (*plugin.Plugin, error) {
	// Load plugins
	l.log.Debugf("Loading plugin %s", path)
	pl, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	return pl, err
}
