package main

import "github.com/kvvPro/metric-collector/tools/certs"

func main() {
	certs.MakeRSACert(&certs.Settings{PathToCert: "/workspaces/metric-collector/cmd/keys/key.pub", PathToPrivateKey: "/workspaces/metric-collector/cmd/keys/key"})
}
