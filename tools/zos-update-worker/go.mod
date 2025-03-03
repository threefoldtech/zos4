module github.com/threefoldtech/zos4/tools/zos-update-version

go 1.23.0

toolchain go1.24.0

require (
	github.com/rs/zerolog v1.28.0
	github.com/spf13/cobra v1.8.1
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
