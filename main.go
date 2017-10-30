package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"flag"

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
	dbUser := flag.String("dbUser", "", "Db user to connect with.")
	dbPass := flag.String("dbPass", "", "Db password to connect with.")
	dbPort := flag.String("dbPort", "3306", "Db port to connect to.")
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

	log.Println("Initialising dbwebapp.")
	userAndPass := *dbUser + ":" + *dbPass
	log.Println("Connecting to " + *dbUser + ":xxxx" + "@tcp(" + *dbHost + ":" + *dbPort + ")/" + *dbName)
	db, err := sql.Open("mysql", userAndPass+"@tcp("+*dbHost+":"+*dbPort+")/"+*dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	numBackOffIterations := 15
	for i := 1; i <= numBackOffIterations; i++ {
		log.Printf("Pinging db %s.\n", *dbHost)
		err = db.Ping()
		if err == nil {
			log.Println("Connected to db.")
			break
		}
		waitDuration := time.Duration(i) * time.Second
		log.Printf("Backing off for %s.\n", waitDuration)
		time.Sleep(waitDuration)
		if i == numBackOffIterations {
			log.Printf("Error connecting to db %s, %s. Exiting.", *dbHost, err)
			os.Exit(1)
		}
	}
	log.Println("Starting dbwebapp server.")
	http.HandleFunc("/health", healthHandler)
	http.ListenAndServe("0.0.0.0:8080", nil)
}
