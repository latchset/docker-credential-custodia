package main

import (
	"fmt"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/latchset/docker-credential-custodia/custodiaservice"
	"os"
)

const defaultSocketPath string = "/var/run/custodia/custodia.sock"
const socketEnvName string = "CUSTODIA_SOCKET"
const defaultBaseURL string = "http://localhost/secrets/docker/"
const baseURLEnvName string = "CUSTODIA_DOCKER_URL"

func main() {
	socketpath := os.Getenv(socketEnvName)
	if socketpath == "" {
		socketpath = defaultSocketPath
	}
	baseurl := os.Getenv(baseURLEnvName)
	if baseurl == "" {
		baseurl = defaultBaseURL
	}
	cs, err := custodiaservice.NewCustodiaService(socketpath, baseurl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	credentials.Serve(cs)
}
