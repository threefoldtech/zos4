package registrargw

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"net/url"
	"sync"
	"time"

	subTypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
	"github.com/threefoldtech/zbus"
	zos4Pkg "github.com/threefoldtech/zos4/pkg"
	"github.com/threefoldtech/zos4/pkg/stubs"
	"github.com/threefoldtech/zosbase/pkg"
	"github.com/threefoldtech/zosbase/pkg/environment"
)

const AuthHeader = "X-Auth"

type registrarGateway struct {
	mu              sync.Mutex
	registrarClient client.RegistrarClient
}

func NewRegistrarGateway(ctx context.Context, cl zbus.Client) (zos4Pkg.RegistrarGateway, error) {
	identity := stubs.NewIdentityManagerStub(cl)
	sk := ed25519.PrivateKey(identity.PrivateKey(ctx))
	hexSeed := hex.EncodeToString(sk.Seed())

	env := environment.MustGet()
	url, err := url.JoinPath(env.RegistrarURL, "v1")
	if err != nil {
		return &registrarGateway{}, err
	}

	cli, err := client.NewRegistrarClient(url, hexSeed)
	if err != nil {
		return &registrarGateway{}, errors.Wrap(err, "failed to create new registrar client")
	}

	gw := &registrarGateway{
		mu:              sync.Mutex{},
		registrarClient: cli,
	}

	return gw, nil
}

func (r *registrarGateway) GetZosVersion() (client.ZosVersion, error) {
	log.Debug().Str("method", "GetZosVersion").Msg("method called")

	return r.registrarClient.GetZosVersion()
}

func (r *registrarGateway) CreateNode(node client.Node) (uint64, error) {
	log.Debug().
		Str("method", "CreateNode").
		Uint32("farm_id", uint32(node.FarmID)).
		Uint32("twin_id", uint32(node.TwinID)).
		Msg("method called")

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.registrarClient.RegisterNode(node.FarmID, node.TwinID, node.Interfaces, node.Location, node.Resources, node.SerialNumber, node.SecureBoot, node.Virtualized)
}

func (r *registrarGateway) CreateTwin(relays []string, rmbEncKey string) (client.Account, error) {
	log.Debug().
		Str("method", "CreateTwin").
		Strs("relay", relays).
		Str("rmbEncKey", rmbEncKey).
		Msg("method called")

	r.mu.Lock()
	defer r.mu.Unlock()

	account, _, err := r.registrarClient.CreateAccount(relays, rmbEncKey)
	return account, err
}

func (r *registrarGateway) EnsureAccount(relays []string, rmbEncKey string) (client.Account, error) {
	log.Debug().
		Str("method", "EnsureAccount").
		Strs("relay", relays).
		Str("rmbEncKey", rmbEncKey).
		Msg("method called")

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.registrarClient.EnsureAccount(relays, rmbEncKey)
}

func (r *registrarGateway) GetFarm(id uint64) (client.Farm, error) {
	log.Debug().
		Str("method", "GetFarm").
		Uint64("farm_id", id).
		Msg("method called")

	return r.registrarClient.GetFarm(id)
}

func (r *registrarGateway) GetNode(id uint64) (client.Node, error) {
	log.Debug().
		Str("method", "GetNode").
		Uint64("node_id", id).
		Msg("method called")

	return r.registrarClient.GetNode(id)
}

func (r *registrarGateway) GetNodeByTwinID(twinID uint64) (client.Node, error) {
	log.Debug().
		Str("method", "GetNodeByTwinID").
		Uint64("twin_id", twinID).
		Msg("method called")

	return r.registrarClient.GetNodeByTwinID(twinID)
}

func (r *registrarGateway) GetNodes(farmID uint64) (nodeIDs []uint64, err error) {
	log.Debug().
		Str("method", "GetNodes").
		Uint64("farm_id", farmID).
		Msg("method called")

	nodes, err := r.registrarClient.ListNodes(client.ListNodesWithFarmID(farmID))
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.NodeID)
	}

	return
}

func (r *registrarGateway) GetTwin(id uint64) (result client.Account, err error) {
	log.Debug().
		Str("method", "GetTwin").
		Uint64("twin_id", id).
		Msg("method called")

	return r.registrarClient.GetAccount(id)
}

func (r *registrarGateway) GetTwinByPubKey(pk []byte) (result uint64, err error) {
	log.Debug().
		Str("method", "GetTwinByPubKey").
		Str("pk", hex.EncodeToString(pk)).
		Msg("method called")

	account, err := r.registrarClient.GetAccountByPK(pk)
	return account.TwinID, err
}

func (r *registrarGateway) UpdateNode(node client.Node) error {
	log.Debug().
		Str("method", "UpdateNode").
		Uint64("twin_id", node.TwinID).
		Msg("method called")

	r.mu.Lock()
	defer r.mu.Unlock()
	opts := []client.UpdateNodeOpts{
		client.UpdateNodesWithFarmID(node.FarmID),
		client.UpdateNodesWithLocation(node.Location),
		client.UpdateNodesWithResources(node.Resources),
		client.UpdateNodesWithInterfaces(node.Interfaces),
		client.UpdateNodesWithSerialNumber(node.SerialNumber),
	}

	if node.SecureBoot {
		opts = append(opts, client.UpdateNodesWithSecureBoot())
	}
	if node.Virtualized {
		opts = append(opts, client.UpdateNodesWithVirtualized())
	}

	return r.registrarClient.UpdateNode(opts...)
}

func (r *registrarGateway) UpdateNodeUptimeV2(uptime time.Duration, timestamp time.Time) (err error) {
	log.Debug().
		Str("method", "UpdateNodeUptimeV2").
		Uint64("uptime", uint64(uptime)).
		Msg("method called")

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.registrarClient.ReportUptime(client.UptimeReport{Uptime: uptime, Timestamp: timestamp})
}

func (r *registrarGateway) GetTime() (time.Time, error) {
	// log.Trace().Str("method", "Time").Msg("method called")
	//
	// return g.sub.Time()
	return time.Now(), nil
}

func (r *registrarGateway) GetContract(id uint64) (result substrate.Contract, serr pkg.SubstrateError) {
	// log.Trace().Str("method", "GetContract").Uint64("id", id).Msg("method called")
	// contract, err := g.sub.GetContract(id)
	//
	// serr = buildSubstrateError(err)
	// if err != nil {
	// 	return
	// }
	return
}

func (r *registrarGateway) GetContractIDByNameRegistration(name string) (result uint64, serr pkg.SubstrateError) {
	// log.Trace().Str("method", "GetContractIDByNameRegistration").Str("name", name).Msg("method called")
	// contractID, err := g.sub.GetContractIDByNameRegistration(name)
	//
	// serr = buildSubstrateError(err)
	return
}

func (r *registrarGateway) GetNodeContracts(node uint32) ([]subTypes.U64, error) {
	// log.Trace().Str("method", "GetNodeContracts").Uint32("node", node).Msg("method called")
	// return g.sub.GetNodeContracts(node)
	return []subTypes.U64{}, nil
}

func (r *registrarGateway) GetNodeRentContract(node uint32) (result uint64, serr pkg.SubstrateError) {
	// log.Trace().Str("method", "GetNodeRentContract").Uint32("node", node).Msg("method called")
	// contractID, err := g.sub.GetNodeRentContract(node)
	//
	// serr = buildSubstrateError(err)
	return
}

func (r *registrarGateway) GetPowerTarget() (power substrate.NodePower, err error) {
	// log.Trace().Str("method", "GetPowerTarget").Uint32("node id", uint32(g.nodeID)).Msg("method called")
	// return g.sub.GetPowerTarget(uint32(g.nodeID))
	power = substrate.NodePower{
		State:  substrate.PowerState{IsUp: true},
		Target: substrate.Power{IsUp: true},
	}

	return
}

func (r *registrarGateway) Report(consumptions []substrate.NruConsumption) (subTypes.Hash, error) {
	// contractIDs := make([]uint64, 0, len(consumptions))
	// for _, v := range consumptions {
	// 	contractIDs = append(contractIDs, uint64(v.ContractID))
	// }
	//
	// log.Debug().Str("method", "Report").Uints64("contract ids", contractIDs).Msg("method called")
	// r.mu.Lock()
	// defer r.mu.Unlock()
	//
	// url := fmt.Sprintf("%s/v1/nodes/%d/consumption", r.baseURL, r.nodeID)
	//
	// var body bytes.Buffer
	// _, err := r.httpClient.Post(url, "application/json", &body)
	// if err != nil {
	// 	return subTypes.Hash{}, err
	// }
	//
	// // I need to know what is hash to be able to respond with it
	// return r.sub.Report(r.identity, consumptions)
	return subTypes.Hash{}, nil
}

func (r *registrarGateway) SetContractConsumption(resources ...substrate.ContractResources) error {
	// contractIDs := make([]uint64, 0, len(resources))
	// for _, v := range resources {
	// 	contractIDs = append(contractIDs, uint64(v.ContractID))
	// }
	// log.Debug().Str("method", "SetContractConsumption").Uints64("contract ids", contractIDs).Msg("method called")
	// g.mu.Lock()
	// defer g.mu.Unlock()
	// return g.sub.SetContractConsumption(g.identity, resources...)
	return nil
}

func (r *registrarGateway) SetNodePowerState(up bool) (hash subTypes.Hash, err error) {
	// log.Debug().Str("method", "SetNodePowerState").Bool("up", up).Msg("method called")
	// g.mu.Lock()
	// defer g.mu.Unlock()
	// return g.sub.SetNodePowerState(g.identity, up)
	return subTypes.Hash{}, nil
}

func buildSubstrateError(err error) (serr pkg.SubstrateError) {
	if err == nil {
		return
	}

	serr.Err = err
	serr.Code = pkg.CodeGenericError

	if errors.Is(err, substrate.ErrNotFound) {
		serr.Code = pkg.CodeNotFound
	} else if errors.Is(err, substrate.ErrBurnTransactionNotFound) {
		serr.Code = pkg.CodeBurnTransactionNotFound
	} else if errors.Is(err, substrate.ErrRefundTransactionNotFound) {
		serr.Code = pkg.CodeRefundTransactionNotFound
	} else if errors.Is(err, substrate.ErrFailedToDecode) {
		serr.Code = pkg.CodeFailedToDecode
	} else if errors.Is(err, substrate.ErrInvalidVersion) {
		serr.Code = pkg.CodeInvalidVersion
	} else if errors.Is(err, substrate.ErrUnknownVersion) {
		serr.Code = pkg.CodeUnknownVersion
	} else if errors.Is(err, substrate.ErrIsUsurped) {
		serr.Code = pkg.CodeIsUsurped
	} else if errors.Is(err, substrate.ErrAccountNotFound) {
		serr.Code = pkg.CodeAccountNotFound
	} else if errors.Is(err, substrate.ErrDepositFeeNotFound) {
		serr.Code = pkg.CodeDepositFeeNotFound
	} else if errors.Is(err, substrate.ErrMintTransactionNotFound) {
		serr.Code = pkg.CodeMintTransactionNotFound
	}
	return
}
