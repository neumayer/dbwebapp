package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/vault/api"
)

type vaultClient struct {
	client *api.Client
	secret *api.Secret
}

func newVaultClient(address string) (*vaultClient, error) {
	config := api.Config{Address: address}
	client, err := api.NewClient(&config)
	if err != nil {
		return nil, err
	}
	return &vaultClient{client: client}, nil
}

func (v *vaultClient) readVaultSecret(path string) error {
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return err
	}
	v.secret = secret
	return nil
}

func (v *vaultClient) getCredentials(vaultAddr, vaultToken, vaultRoleID, vaultSecretID string) (string, string, error) {
	if vaultRoleID != "" && vaultSecretID != "" {
		options := map[string]interface{}{
			"role_id":   vaultRoleID,
			"secret_id": vaultSecretID,
		}
		path := "auth/approle/login"
		pingExternalService(vaultAddr, &vaultAppRolePinger{v.client, path, options})
		secret, err := v.client.Logical().Write(path, options)
		if err != nil {
			return "", "", err
		}
		v.client.SetToken(secret.Auth.ClientToken)
		if secret.Auth == nil {
			return "", "", fmt.Errorf("could not read auth info from secret")
		}
		err = v.readVaultSecret("database/creds/vault-mysql-role")
		if err != nil {
			return "", "", err
		}
		username := v.secret.Data["username"].(string)
		password := v.secret.Data["password"].(string)
		return username, password, nil
	}
	if vaultToken != "" {
		v.client.SetToken(vaultToken)
		path := "secrets/dbwebapp"
		pingExternalService(vaultAddr, &vaultPinger{v.client, path})
		err := v.readVaultSecret(path)
		if err != nil {
			return "", "", err
		}
		username := v.secret.Data["username"].(string)
		password := v.secret.Data["password"].(string)
		return username, password, nil
	}
	return "", "", fmt.Errorf("could not read vault secret")
}

func (v *vaultClient) renewLease() error {
	log.Printf("Renewing lease %v.", v.secret.LeaseID)
	_, err := v.client.Sys().Renew(v.secret.LeaseID, v.secret.LeaseDuration)
	if err != nil {
		return err
	}
	return nil
}

func (v *vaultClient) regularlyRenewLease() error {
	if !v.secret.Renewable {
		log.Println("Cowardly refusing to renew unrenewable secret.")
		return nil
	}
	v.renewLease()
	// renew lease 100 seconds before expiry
	interval := time.Duration(v.secret.LeaseDuration)*time.Second - 100*time.Second
	renewTicker := time.NewTicker(interval)
	log.Printf("Scheduling regular renewal for lease %s every %v", v.secret.LeaseID, interval)

	for {
		select {
		case <-renewTicker.C:
			v.renewLease()
		}
	}
}
