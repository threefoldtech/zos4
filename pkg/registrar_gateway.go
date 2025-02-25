package pkg

import (
	"time"

	substrateTypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
	"github.com/threefoldtech/zos4/pkg/types"
	"github.com/threefoldtech/zosbase/pkg"
)

//go:generate zbusc -module api-gateway -version 0.0.1 -name api-gateway -package stubs github.com/threefoldtech/zos4/pkg+RegistrarGateway stubs/registrar-gateway.go

type RegistrarGateway interface {
	CreateTwin(relay []string, rmbEncKey string) (client.Account, error)
	EnsureAccount(relay []string, rmbEncKey string) (twin client.Account, err error)
	GetTwin(id uint64) (client.Account, error)
	GetTwinByPubKey(pk []byte) (uint64, error)

	CreateNode(node types.UpdateNodeRequest) (uint64, error)
	GetNode(id uint64) (client.Node, error)
	GetNodes(farmID uint64) ([]uint64, error)
	GetNodeByTwinID(twin uint64) (client.Node, error)
	UpdateNode(node types.UpdateNodeRequest) error
	UpdateNodeUptimeV2(uptime time.Duration, timestamp time.Time) (err error)

	GetFarm(id uint64) (client.Farm, error)

	GetTime() (time.Time, error)
	GetZosVersion() (client.ZosVersion, error)

	GetNodeContracts(node uint32) ([]substrateTypes.U64, error)
	GetContract(id uint64) (substrate.Contract, pkg.SubstrateError)
	GetContractIDByNameRegistration(name string) (uint64, pkg.SubstrateError)
	GetNodeRentContract(node uint32) (uint64, pkg.SubstrateError)
	GetPowerTarget() (power substrate.NodePower, err error)
	Report(consumptions []substrate.NruConsumption) (substrateTypes.Hash, error)
	SetContractConsumption(resources ...substrate.ContractResources) error
	SetNodePowerState(up bool) (hash substrateTypes.Hash, err error)
}
