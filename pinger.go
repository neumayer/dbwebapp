package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/hashicorp/vault/api"
)

type pinger interface {
	ping() error
}

type dbPinger struct {
	db *sql.DB
}

func (d dbPinger) ping() error {
	return d.db.Ping()
}

// pingExternalService pings an external service with linearly increasing backoff time.
func pingExternalService(addr string, pinger pinger) error {
	numBackOffIterations := 15
	for i := 1; i <= numBackOffIterations; i++ {
		log.Printf("Pinging %s.\n", addr)
		err := pinger.ping()
		if err != nil {
			log.Println(err)
		}
		if err == nil {
			log.Printf("Connected to %s.", addr)
			break
		}
		waitDuration := time.Duration(i) * time.Second
		log.Printf("Backing off for %v.\n", waitDuration)
		time.Sleep(waitDuration)
		if i == numBackOffIterations {
			return err
		}
	}
	return nil
}

type vaultPinger struct {
	vaultClient *api.Client
	path        string
}

func (v *vaultPinger) ping() error {
	_, err := v.vaultClient.Logical().Read(v.path)
	return err
}

type vaultAppRolePinger struct {
	vaultClient *api.Client
	path        string
	options     map[string]interface{}
}

func (v *vaultAppRolePinger) ping() error {
	_, err := v.vaultClient.Logical().Write(v.path, v.options)
	return err
}
