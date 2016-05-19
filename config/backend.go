package config

import "errors"

var (
	PublicKeyNotFoundErr = errors.New("Public key not found")
)

type Backend interface {
	LoadConfig(filepath string) (Configuration, error)

	Keys() ([]PublicKey, error)
	AddKey(PublicKey) error
	GetKey(fingerprint string) (PublicKey, error)
	DeleteKey(fingerprint string) error
}

type PublicKey struct {
	Key         []byte `json:"key"`
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment"`
}

type Configuration struct {
	RPC               string `json:"rpc"`
	SSH               string `json:"ssh"`
	Hostname          string `json:"hostname"`
	NotificationEmail string `json:"notificationEmail"`
	PrivateRegistry   string `json:"privateRegistry"`
	GitPath           string `json:"gitpath"`
	DockerSock        string `json:"dockersock"`
	DataDir           string `json:"dataDir"`
	MasterOnly        bool   `json:"masterOnly"`
	Debug             bool   `json:"debug"`
}

type BackendFactory func(url string) (Backend, error)
