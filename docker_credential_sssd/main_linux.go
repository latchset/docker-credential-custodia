package main

import (
	"fmt"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/latchset/docker-credential-custodia/custodiaservice"
	"os"
)

const defaultSocketPath string = "/var/run/secrets.socket"
const defaultBaseURL string = "http://localhost/secrets/docker/"
const baseURLEnvName string = "SSSD_SECRETS_DOCKER_URL"

func main() {
	baseurl := os.Getenv(baseURLEnvName)
	if baseurl == "" {
		baseurl = defaultBaseURL
	}
	cs, err := custodiaservice.NewCustodiaService(defaultSocketPath, baseurl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	credentials.Serve(cs)
}
