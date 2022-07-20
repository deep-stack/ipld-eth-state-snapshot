package fixture

import (
	"os"
	"path/filepath"
)

// TODO: embed some mainnet data
// import "embed"
//_go:embed mainnet_data.tar.gz

func GetChainDataPath(path string) (string, string) {
	chaindataPath, err := filepath.Abs(path)
	if err != nil {
		panic("cannot resolve path " + path)
	}
	ancientdataPath := filepath.Join(chaindataPath, "ancient")

	if _, err := os.Stat(chaindataPath); err != nil {
		panic("must populate chaindata at " + chaindataPath)
	}

	return chaindataPath, ancientdataPath
}
