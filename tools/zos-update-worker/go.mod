module github.com/threefoldtech/zos4/tools/zos-update-version

go 1.23.0

toolchain go1.24.0

require (
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.33.0
	github.com/spf13/cobra v1.8.1
	github.com/threefoldtech/tfgrid4-sdk-go/node-registrar v0.0.0-20250409145235-3ffee627d2ae
)

require (
	github.com/ChainSafe/go-schnorrkel v1.0.0 // indirect
	github.com/cosmos/go-bip39 v1.0.0 // indirect
	github.com/decred/base58 v1.0.4 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/gtank/merlin v0.1.1 // indirect
	github.com/gtank/ristretto255 v0.1.2 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20220103164710-9a04d6ca976b // indirect
	github.com/vedhavyas/go-subkey/v2 v2.0.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
)

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace github.com/centrifuge/go-substrate-rpc-client/v4 v4.0.5 => github.com/threefoldtech/go-substrate-rpc-client/v4 v4.0.6-0.20220927094755-0f0d22c73cc7
