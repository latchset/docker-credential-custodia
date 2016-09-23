# docker-credential-custodia

Store Docker credentials in Custodia or SSSD.

Authors:
  * Christian Heimes <cheimes@redhat.com>

References:
  * https://github.com/latchset/custodia
  * https://github.com/docker/docker-credential-helpers
  * https://fedorahosted.org/sssd/wiki/DesignDocs/SecretsService
  * http://www.projectatomic.io/blog/2016/03/docker-credentials-store/

## Requirements

  * Docker >= 1.11
  * SSSD >= 1.14 (for docker-credential-sssd)
  * Custodia >= 0.2 (docker-credential-custodia)
  * golang

## Build

```
export GOPATH=~/godeps
mkdir -p $GOPATH/src/github.com/latchset
git clone ... $GOPATH/src/github.com/latchset/docker-credential-custodia
cd $GOPATH/src/github.com/latchset/docker-credential-custodia
make
```

## Use with SSSD

Start SSSD secret responder

```
systemctl enable sssd-secrets.socket
systemctl start sssd-secrets.socket
```

Copy helper to /usr/bin
```
cp docker-credential-sssd /usr/bin/
```

Enable SSSD credential store
```
mkdir -p ~/.docker
cp ~/.docker/config.json ~/.docker/config.json.bak
echo '{"credsStore": "sssd"}' > ~/.docker/config.json
```

## License

[MIT](LICENSE)
