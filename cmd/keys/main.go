package main

import "github.com/kvvPro/metric-collector/tools/encrypt"

func main() {
	encrypt.MakeRSACert(&encrypt.Settings{PathToCert: "/workspaces/metric-collector/cmd/keys/key.pub", PathToPrivateKey: "/workspaces/metric-collector/cmd/keys/key"})
}
