package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync/atomic"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/kokizzu/gotro/L"
	"github.com/kokizzu/gotro/S"

	"vaultdist1/cmd/vault-plugin-dnskey/dnskey"
)

const vaultAddr = `http://127.0.0.1:8200`
const writerApproleId = `writer1_approle1`
const writerTokenFile = `/tmp/writer1-secret`
const readerApproleId = `reader1_approle1`
const readerTokenFile = `/tmp/reader1-secret`
const vaultPathPrefix = `secret/data/keys1/`
const vaultDnskeyPathPrefix = `mysecret/dnskey/keys2/`

func main() {
	if len(os.Args) == 1 {
		fmt.Println(`Usage: 
go run main.go write zoneName # also overwrite if exists, see version for revision 
go run main.go read zoneName
go run main.go delete zoneName
go run main.go benchmark
go run main.go writeplugin zoneName # also overwrite if exists, see version for revision 
go run main.go readplugin zoneName
go run main.go deleteplugin zoneName
go run main.go benchmarkplugin
`)
		os.Exit(0)
	}

	switch os.Args[1] {
	case `writeplugin`:
		var zoneName string
		if len(os.Args) > 2 {
			zoneName = os.Args[2]
		} else {
			zoneName = S.RandomPassword(10) + `.com`
			fmt.Println(`no zoneName provided, using random: ` + zoneName)
		}

		// connect as writer1
		vaultClient := createVaultClient(vaultAddr, writerApproleId, writerTokenFile)

		vaultPath := vaultDnskeyPathPrefix + zoneName
		res, err := vaultClient.Logical().Write(vaultPath, map[string]any{
			`zone`: zoneName,
		})
		L.PanicIf(err, `client.Logical().Write: `+vaultPath)
		L.Print(res)
	case `readplugin`:
		if len(os.Args) < 2 {
			fmt.Println(`Usage: go run main.go readplugin zoneName`)
			os.Exit(1)
		}
		zoneName := os.Args[2]

		// connect as reader1
		vaultClient := createVaultClient(vaultAddr, readerApproleId, readerTokenFile)

		vaultPath := vaultDnskeyPathPrefix + zoneName
		secret, err := vaultClient.Logical().Read(vaultPath)
		L.PanicIf(err, `client.Logical().Write: `+vaultPath)

		if secret.Data != nil {
			L.Describe(secret.Data)
		}

	case `deleteplugin`:
		if len(os.Args) < 2 {
			fmt.Println(`Usage: go run main.go deleteplugin zoneName`)
			os.Exit(1)
		}
		zoneName := os.Args[2]

		// connect as writer1
		vaultClient := createVaultClient(vaultAddr, writerApproleId, writerTokenFile)

		vaultPath := vaultPathPrefix + zoneName
		_, err := vaultClient.Logical().Delete(vaultPath)
		L.PanicIf(err, `client.Logical().Delete: `+vaultPath)

	case `write`:
		var zoneName string
		if len(os.Args) > 2 {
			zoneName = os.Args[2]
		} else {
			zoneName = S.RandomPassword(10) + `.com`
			fmt.Println(`no zoneName provided, using random: ` + zoneName)
		}
		v := dnskey.GenerateDNSKey(zoneName)

		// connect as writer1
		vaultClient := createVaultClient(vaultAddr, writerApproleId, writerTokenFile)

		vaultPath := vaultPathPrefix + zoneName
		res, err := vaultClient.Logical().Write(vaultPath, map[string]any{
			`data`: v.ToMap(), // have to have "data"
		})
		L.PanicIf(err, `client.Logical().Write: `+vaultPath)
		L.Print(res)
	case `read`:
		if len(os.Args) < 2 {
			fmt.Println(`Usage: go run main.go read zoneName`)
			os.Exit(1)
		}
		zoneName := os.Args[2]

		// connect as reader1
		vaultClient := createVaultClient(vaultAddr, readerApproleId, readerTokenFile)

		vaultPath := vaultPathPrefix + zoneName
		secret, err := vaultClient.Logical().Read(vaultPath)
		L.PanicIf(err, `client.Logical().Write: `+vaultPath)

		if secret.Data != nil {
			L.Describe(secret.Data[`metadata`])
			L.Describe(secret.Data[`data`])
		}

	case `delete`:
		if len(os.Args) < 2 {
			fmt.Println(`Usage: go run main.go delete zoneName`)
			os.Exit(1)
		}
		zoneName := os.Args[2]

		// connect as writer1
		vaultClient := createVaultClient(vaultAddr, writerApproleId, writerTokenFile)

		vaultPath := vaultPathPrefix + zoneName
		_, err := vaultClient.Logical().Delete(vaultPath)
		L.PanicIf(err, `client.Logical().Delete: `+vaultPath)

	case `benchmark`:
		var writeCounter, readCounter, errCounter uint64
		start := time.Now()
		go writerProcess(&writeCounter, &errCounter)
		go readerProcess(&readCounter, &errCounter)
		go readerProcess(&readCounter, &errCounter)
		go readerProcess(&readCounter, &errCounter)
		go readerProcess(&readCounter, &errCounter)
		for {
			dur := time.Since(start).Seconds()
			fmt.Printf("\rwriteCounter: %d (%.2f/s), readCounter: %d (%.2f/s), err: %d",
				writeCounter,
				float64(writeCounter)/dur,
				readCounter,
				float64(readCounter)/dur,
				errCounter,
			)
			time.Sleep(200 * time.Millisecond)
		}
	case `benchmarkplugin`:
		var writeCounter, readCounter, errCounter uint64
		start := time.Now()
		go writerPluginProcess(&writeCounter, &errCounter)
		go readerPluginProcess(&readCounter, &errCounter)
		go readerPluginProcess(&readCounter, &errCounter)
		go readerPluginProcess(&readCounter, &errCounter)
		go readerPluginProcess(&readCounter, &errCounter)
		for {
			dur := time.Since(start).Seconds()
			fmt.Printf("\rwriteCounter: %d (%.2f/s), readCounter: %d (%.2f/s), err: %d",
				writeCounter,
				float64(writeCounter)/dur,
				readCounter,
				float64(readCounter)/dur,
				errCounter,
			)
			time.Sleep(200 * time.Millisecond)
		}
	}

}

var zoneList []string

func readerProcess(u *uint64, errCount *uint64) {
	vaultReader := createVaultClient(vaultAddr, readerApproleId, readerTokenFile)

	for {
		if len(zoneList) == 0 {
			time.Sleep(time.Millisecond)
			continue
		}
		zoneName := zoneList[rand.Int()%len(zoneList)]

		vaultPath := vaultPathPrefix + zoneName
		secret, err := vaultReader.Logical().Read(vaultPath)
		if err != nil || secret == nil || secret.Data == nil {
			atomic.AddUint64(errCount, 1)
			if err != nil {
				log.Println("\n", err)
			}
		} else {
			atomic.AddUint64(u, 1)
		}
	}
}

func writerProcess(ok *uint64, errCount *uint64) {
	vaultWriter := createVaultClient(vaultAddr, writerApproleId, writerTokenFile)

	for {
		zoneName := S.RandomPassword(8) + `.com`
		dnsKey := dnskey.GenerateDNSKey(zoneName)
		zoneList = append(zoneList, zoneName)

		vaultPath := vaultPathPrefix + zoneName
		_, err := vaultWriter.Logical().Write(vaultPath, map[string]any{
			`data`: dnsKey.ToMap(),
		})
		if err != nil {
			atomic.AddUint64(errCount, 1)
			log.Println("\n", err)
		} else {
			atomic.AddUint64(ok, 1)
		}
	}
}

func readerPluginProcess(u *uint64, errCount *uint64) {
	vaultReader := createVaultClient(vaultAddr, readerApproleId, readerTokenFile)

	for {
		if len(zoneList) == 0 {
			time.Sleep(time.Millisecond)
			continue
		}
		zoneName := zoneList[rand.Int()%len(zoneList)]

		vaultPath := vaultDnskeyPathPrefix + zoneName
		secret, err := vaultReader.Logical().Read(vaultPath)
		if err != nil || secret == nil || secret.Data == nil {
			atomic.AddUint64(errCount, 1)
			if err != nil {
				log.Println("\n", err)
			}
		} else {
			atomic.AddUint64(u, 1)
		}
	}
}

func writerPluginProcess(ok *uint64, errCount *uint64) {
	vaultWriter := createVaultClient(vaultAddr, writerApproleId, writerTokenFile)

	for {
		zoneName := S.RandomPassword(8) + `.com`
		zoneList = append(zoneList, zoneName)

		vaultPath := vaultDnskeyPathPrefix + zoneName
		_, err := vaultWriter.Logical().Write(vaultPath, map[string]any{
			`zone`: zoneName,
		})
		if err != nil {
			atomic.AddUint64(errCount, 1)
			log.Println("\n", err)
		} else {
			atomic.AddUint64(ok, 1)
		}
	}
}
func createVaultClient(vaultAddr, approleId, tokenFile string) *vault.Client {
	config := vault.DefaultConfig()
	config.Address = vaultAddr
	vaultClient, err := vault.NewClient(config)
	L.PanicIf(err, `vault.NewClient`)
	//err = config.ConfigureTLS(&vault.TLSConfig{Insecure: true}) // not needed for localhost, but required if https cert doesn't contain proper domain name
	//L.PanicIf(err, `config.ConfigureTLS`)

	approleAuth, err := approle.NewAppRoleAuth(approleId, &approle.SecretID{
		FromFile: tokenFile,
	})
	L.PanicIf(err, `approle.NewAppRoleAuth`)

	_, err = vaultClient.Auth().Login(context.Background(), approleAuth)
	L.PanicIf(err, `vaultClient.Auth().Login`)

	return vaultClient
}
