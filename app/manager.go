package app

import (
	"errors"
	"io"
	"time"

	"github.com/adamveld12/goku"
)

const (
	Stopped = Status("Stopped")
	Running = Status("Running")
)

var (
	ErrAppNotExist = errors.New("App by specified name does not exist")
)

// Status is the app's current status
type Status string

// Manager manages apps running under Goku
type Manager interface {
	Get(repository, commit string) (Header, error)
	// For running images -
	//Run(name, image string) (Header, error)
	Run(repository, commit string) (Header, error)
	Kill(repository, commit string) error
	List(filter string) ([]Header, error)
}

// Header holds info about an app that is running in Goku
type Header struct {
	// The URL friendly name for this App
	Name string
	// Repository is the url path to the git repo that this App was built from
	Repository string
	// Commit is the branch/commit hash that this App was built from
	Commit string
	// URL is the url name to access this service
	URL string
	// ContainerID the
	ContainerID string
	// Status is a human readable status of this
	Status    string
	StartTime time.Time
}

// New initializes and returns a new instance of Manager
func New(backend goku.Backend, config goku.Configuration, output io.Writer) Manager {
	return &dockerManager{
		goku.NewLog("[docker appman]"),
		output,
		config,
		backend,
	}
}
