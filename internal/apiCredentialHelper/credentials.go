package apiCredentialHelper

import (
	"bbbstatus/internal/config"
	db "bbbstatus/internal/database"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// big blue button server may change their shared secret during the runtime of bbbstatus. we keep the credentials of a big blue button server in this credentials store, which is a sync map, which maps from a hostname to a big blue button server or a scale lite server and falls back to the config credentials or the database credentials.
var credentialStoreBBBServers = new(sync.Map)

func LoadCredentialsFromDatabase(conf config.ConfigurationStruct) error {
	var ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancel()

	conn, err := pgx.Connect(ctx, conf.DatabaseConfig.DatabaseConnectionString)
	if err != nil {
		fmt.Println("error occurred during pgx connect (bbbWebHookEvent): ", err)
		return errors.New("error occurred database connect internal/apiCredentialHelper/LoadCredentialsFromDatabase")
	}
	defer conn.Close(ctx)

	dbQueries := db.New(conn)
	for _, server := range conf.BBBServers {
		var dbServer db.BbbServer
		dbServer, err = dbQueries.GetBBBServerByHostname(ctx, server.Hostname)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				credentialStoreBBBServers.Store(server.Hostname, server)
				continue
			}
			fmt.Println("error occurred during getBBBServerByHostname (query database for bbbserver): ", err)
			return errors.New("error occurred database connect internal/apiCredentialHelper/LoadCredentialsFromDatabase")
		}
		credentialStoreBBBServers.Store(server.Hostname, dbServer)
	}
	return nil
}

func GetApiKey(hostname string) (apikey string) {
	var bbbServer config.BbbServer
	var scaleLiteServer config.ScaleliteServer
	server, ok := credentialStoreBBBServers.Load(hostname)
	if !ok {
		return
	}

	bbbServer, ok = server.(config.BbbServer)
	if ok {
		return bbbServer.SharedSecret
	}
	scaleLiteServer, ok = server.(config.ScaleliteServer)
	if ok {
		return scaleLiteServer.SharedSecret
	}
	return
}

func UpdateApiKey(hostname string, apikey string) error {
	var bbbServer config.BbbServer
	var bbbServerLoadOK bool
	var server interface{}

	server, bbbServerLoadOK = credentialStoreBBBServers.Load(hostname)
	bbbServer = server.(config.BbbServer)
	if !bbbServerLoadOK {
		return fmt.Errorf("credentialStoreBBBServers.Load() failed, cannot find credentials in  bbb server store")
	}

	bbbServer.SharedSecret = apikey
	credentialStoreBBBServers.Store(hostname, bbbServer)
	return nil
}
