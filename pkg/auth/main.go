package auth

import (
	"bytes"
	"fmt"
	"net/http"

	auth0 "github.com/auth0-community/go-auth0"
	jose "gopkg.in/square/go-jose.v2"
)

func main() {
	const iss = "https://ledger-staging.eu.auth0.com/"
	client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: iss + ".well-known/jwks.json"}, nil)
	// audience := os.Getenv("AUTH0_CLIENT_ID")
	audience := "https://staging.api.my-ledger.com"
	configuration := auth0.NewConfiguration(client, []string{audience}, iss, jose.RS256)
	validator := auth0.NewValidator(configuration, nil)

	req, _ := http.NewRequest("POST", "/v1/hello", bytes.NewBufferString("Some crap"))
	// req.Header.Add("Authorization", "BEARER eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6IlJrSTFPVEEzTVVSRk56WkZPVUZDUWpaRE16TXpPVEpCTlRjMFFVTTRSakl5UWpBNE9URXlSZyJ9.eyJpc3MiOiJodHRwczovL2xlZGdlci1zdGFnaW5nLmV1LmF1dGgwLmNvbS8iLCJzdWIiOiI5T2ZXTHhYT2t4Zkd1bE5oY2tmNnJ2S0ZjZnI0azNqY0BjbGllbnRzIiwiYXVkIjoiaHR0cHM6Ly9zdGFnaW5nLmFwaS5teS1sZWRnZXIuY29tIiwiaWF0IjoxNTI4MzE4ODU3LCJleHAiOjE1Mjg0MDUyNTcsImF6cCI6IjlPZldMeFhPa3hmR3VsTmhja2Y2cnZLRmNmcjRrM2pjIiwic2NvcGUiOiJ3cml0ZTphY2NvdW50LWNhdGVnb3JpZXMgcmVhZDphY2NvdW50LWNhdGVnb3JpZXMgd3JpdGU6dGFncyByZWFkOnRhZ3Mgd3JpdGU6dHJhbnNhY3Rpb25zIHJlYWQ6dHJhbnNhY3Rpb25zIHdyaXRlOmxlZGdlcnMgcmVhZDpsZWRnZXJzIHdyaXRlOmFjY291bnRzIHJlYWQ6YWNjb3VudHMiLCJndHkiOiJjbGllbnQtY3JlZGVudGlhbHMifQ.bKFdmbBW_71MjiinB2g4Y1S58n_mVvFp66azzcQdkUVS09vxGUvvhTavLpFnoaFV6vr7FJa7kPxweQEzANlRrqikJkViXD6aj8RVBLnfp7ljROQ_hZr6u22BpsQ_8hJyRVbM-Ddd1VM2waWWN4V_9V5hJRyue8OideJkFxCAhT-hxuFG6IuAWQ3bwu9IPH15lXfc0KPFopbswh-YtGhAu0RnbhhHsd20Ga-3iiyvlg5NRhRU4hFRpMS0up7yw8tZie1RbpXjwpxH5UwXxsIA3qWIKMXSTsyNCs-GOwsXOgXcolkCaExuWKfMP4O7SJInuo8eGZ39koWXjT4Q0m_J6Q")

	token, err := validator.ValidateRequest(req)

	if err != nil {
		fmt.Println("Token is not valid:", token)
		fmt.Println(err)
		return
	}

	fmt.Println("Token is valid:", token)

	claims := map[string]interface{}{}

	validator.Claims(req, token, &claims)

	fmt.Println(claims)
}
