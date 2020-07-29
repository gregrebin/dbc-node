package cmd

import (
	"dbc-node/modules"
	"encoding/hex"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var genUsers = map[string]int64{
	"0476bf074f9f881b24619c3ffbb33683069626f117f3a3fd2f1ddda13b3485b45cd2e084356d0571fa370b33ec770039c6a6b371ae4b37ac99a8d708ed3b38d3fc": modules.SatsSupply / 10,
}
var genValidators map[string]int64 = map[string]int64{
	"c468322724705d01fe22c6727890a9a9293d006bc873e73342d85fb36716642c": modules.ToSats(10),
}
var genValidatorKeys []ed25519.PubKeyEd25519

func init() {
	for key := range genValidators {
		var validatorKey ed25519.PubKeyEd25519
		bytes, _ := hex.DecodeString(key)
		copy(validatorKey[:], bytes)
		genValidatorKeys = append(genValidatorKeys, validatorKey)
	}
}
