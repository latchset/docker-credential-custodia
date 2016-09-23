BASEPROJECT=github.com/latchset/docker-credential-custodia
# omit DWARF, symbol table and debug infos to get a smaller binary
BUILDFLAGS=-ldflags="-s -w"

EXAMPLE_APP=./docker-credential-custodia

.PHONY=all
all: custodia sssd

.PHONY=goget
goget:
	go get $(BASEPROJECT)/docker_credential_custodia

.PHONY=custodia
custodia: goget
	go build $(BUILDFLAGS) -o docker-credential-custodia $(BASEPROJECT)/docker_credential_custodia

.PHONY=sssd
sssd: goget
	go build $(BUILDFLAGS) -o docker-credential-sssd $(BASEPROJECT)/docker_credential_sssd

.PHONY=fmt
fmt:
	go fmt $(BASEPROJECT)/custodiaservice
	go fmt $(BASEPROJECT)/docker_credential_custodia
	go fmt $(BASEPROJECT)/docker_credential_sssd

.PHONY=example
example: custodia
	curl -s --unix-socket /var/run/custodia/custodia.sock -XDELETE http://localhost/secrets/docker/ || true
	$(EXAMPLE_APP) list
	echo '{"ServerURL": "http://localhost:5000", "Username": "user", "Secret": "password"}' | $(EXAMPLE_APP) store
	$(EXAMPLE_APP) list
	echo "http://localhost:5000" | $(EXAMPLE_APP) get
	echo '{"ServerURL": "http://otherhost:5000", "Username": "other", "Secret": "secret"}' | $(EXAMPLE_APP) store
	$(EXAMPLE_APP) list
	echo "http://localhost:5000" | $(EXAMPLE_APP) erase
	$(EXAMPLE_APP) list
	echo "http://otherhost:5000" | $(EXAMPLE_APP) erase
	echo "http://localhost:5000" | $(EXAMPLE_APP) get || true
