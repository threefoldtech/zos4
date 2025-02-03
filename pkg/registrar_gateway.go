package pkg

import (
	"time"

	substrateTypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/zos4/pkg/types"
	"github.com/threefoldtech/zosbase/pkg"
)

//go:generate zbusc -module api-gateway -version 0.0.1 -name api-gateway -package stubs github.com/threefoldtech/zos4/pkg+RegistrarGateway stubs/registrar-gateway.go

type RegistrarGateway interface {
	CreateNode(node types.UpdateNodeRequest) (uint64, error)
	CreateTwin(relay string, pk []byte) (types.Account, error)
	EnsureAccount(pk []byte) (twin types.Account, err error)
	GetContract(id uint64) (substrate.Contract, pkg.SubstrateError)
	GetContractIDByNameRegistration(name string) (uint64, pkg.SubstrateError)
	GetNodeRentContract(node uint32) (uint64, pkg.SubstrateError)
	GetFarm(id uint64) (types.Farm, error)
	GetNode(id uint64) (types.Node, error)
	GetNodeByTwinID(twin uint64) (uint64, error)
	GetNodeContracts(node uint32) ([]substrateTypes.U64, error)
	GetNodes(farmID uint32) ([]uint32, error)
	GetPowerTarget() (power substrate.NodePower, err error)
	GetTwin(id uint64) (types.Account, error)
	GetTwinByPubKey(pk []byte) (uint64, error)
	Report(consumptions []substrate.NruConsumption) (substrateTypes.Hash, error)
	SetContractConsumption(resources ...substrate.ContractResources) error
	SetNodePowerState(up bool) (hash substrateTypes.Hash, err error)
	UpdateNode(node types.UpdateNodeRequest) (uint64, error)
	UpdateNodeUptimeV2(uptime uint64, timestampHint uint64) (err error)
	GetTime() (time.Time, error)
	GetZosVersion() (string, error)
}
