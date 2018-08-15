package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// healthHandler returns http status ok.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alive!"))
}

// parseStringEnv parses an environment variable into a string for both the
// name itself and an uppercase version of the name (e.g. for a flag name of
// "dbHost" we try both "dbHost" and "DBHOST" to be a bit flexible).
func parseStringEnv(flagName string) string {
	if os.Getenv(flagName) != "" {
		return os.Getenv(flagName)
	}
	return os.Getenv(strings.ToUpper(flagName))
}

// main parses input parameters (or reads them from the environment,
// environment variables take precedence), constructs a db connection string,
// tries to connect to the given db, and exposes a simple health endpoint on
// success.
func main() {
	dbHost := flag.String("dbHost", "localhost", "Db host to connect to.")
	dbName := flag.String("dbName", "", "Db to use.")
	dbUser := flag.String("dbUser", "", "Db user to connect with. Requires dbPass. Takes precedence over vault methods")
	dbPass := flag.String("dbPass", "", "Db password to connect with. Requires dbUser. Takes precedence over vault methods")
	dbPort := flag.String("dbPort", "3306", "Db port to connect to.")
	vaultToken := flag.String("vaultToken", "", "Token to access vault with. Takes precedence over vaultRoleId and vaultSecredId.")
	vaultRoleID := flag.String("vaultRoleId", "", "Role ID to access vault with. Requires vaultSecretId.")
	vaultSecretID := flag.String("vaultSecretId", "", "Secret ID to access vault with. Requires vaultRoleId.")
	vaultAddr := flag.String("vaultAddr", "http://localhost:8200", "Vault address to use.")
	flag.Parse()
	envDbHost := parseStringEnv("dbHost")
	if envDbHost != "" {
		dbHost = &envDbHost
	}
	envDbName := parseStringEnv("dbName")
	if envDbName != "" {
		dbName = &envDbName
	}
	envDbUser := parseStringEnv("dbUser")
	if envDbUser != "" {
		dbUser = &envDbUser
	}
	envDbPass := parseStringEnv("dbPass")
	if envDbPass != "" {
		dbPass = &envDbPass
	}
	envDbPort := parseStringEnv("dbPort")
	if envDbPort != "" {
		dbPort = &envDbPort
	}

	envVaultToken := parseStringEnv("vaultToken")
	if envVaultToken != "" {
		vaultToken = &envVaultToken
	}

	envVaultSecretID := parseStringEnv("vaultSecretID")
	if envVaultSecretID != "" {
		vaultSecretID = &envVaultSecretID
	}

	envVaultRoleID := parseStringEnv("vaultRoleId")
	if envVaultRoleID != "" {
		vaultRoleID = &envVaultRoleID
	}

	envVaultAddr := parseStringEnv("vaultAddr")
	if envVaultAddr != "" {
		vaultAddr = &envVaultAddr
	}

	log.Println("Initialising dbwebapp.")
	errChan := make(chan error)

	// get db credentials
	if *dbUser == "" && *dbPass == "" && (*vaultToken != "" || (*vaultRoleID != "" && *vaultSecretID != "")) {
		vaultClient, err := newVaultClient(*vaultAddr)
		*dbUser, *dbPass, err = vaultClient.getCredentials(*vaultAddr, *vaultToken, *vaultRoleID, *vaultSecretID)
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			errChan <- vaultClient.regularlyRenewLease()
		}()
	}
	if *dbUser == "" || *dbPass == "" {
		log.Fatal("No database credentials given (neither via environment, command line, or vault)")
	}

	// set up db
	userAndPass := *dbUser + ":" + *dbPass
	log.Println("Connecting to " + *dbUser + ":xxxx" + "@tcp(" + *dbHost + ":" + *dbPort + ")/" + *dbName)
	db, err := sql.Open("mysql", userAndPass+"@tcp("+*dbHost+":"+*dbPort+")/"+*dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = pingExternalService(*dbHost, &dbPinger{db})
	if err != nil {
		log.Fatal(err)
	}

	// start http server
	log.Println("Starting dbwebapp server.")
	http.HandleFunc("/health", healthHandler)
	go func() {
		errChan <- http.ListenAndServe("0.0.0.0:8081", nil)
	}()
	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
