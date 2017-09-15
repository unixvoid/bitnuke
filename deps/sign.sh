 #!/bin/bash
 echo $1 | gpg \
	--passphrase-fd 0 \
	--batch --yes \
	--no-default-keyring --armor \
	--secret-keyring ./unixvoid.sec --keyring ./unixvoid.pub \
	--output bitnuke-api-0.20.1-linux-amd64.aci.asc \
	--detach-sig bitnuke-api-0.20.1-linux-amd64.aci
