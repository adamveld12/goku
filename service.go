package goku

import (
	"log"
	"os"
	"os/signal"
)

var services []ServiceConstructor

type ServiceConstructor func(Configuration, Backend) Service

func RegisterService(sc ServiceConstructor) {
	services = append(services, sc)
}

func StartServices(config Configuration) error {
	log.Println("starting services")
	backend, _ := NewBackend("debug", "")
	for _, sc := range services {
		s := sc(config, backend)

		if err := s.Start(); err != nil {
			return err
		}

		defer s.Stop()
	}

	log.Println("waiting for signal")
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	signal.Stop(sigs)
	return nil
}

type Service interface {
	Start() error
	Stop() error
}
