package registrar

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
	"github.com/threefoldtech/zbus"
	zos4Stubs "github.com/threefoldtech/zos4/pkg/stubs"
	"github.com/threefoldtech/zosbase/pkg/environment"
	"github.com/threefoldtech/zosbase/pkg/geoip"
	gridtypes "github.com/threefoldtech/zosbase/pkg/gridtypes"
	"github.com/threefoldtech/zosbase/pkg/stubs"
)

type RegistrationInfo struct {
	Capacity     gridtypes.Capacity
	Location     geoip.Location
	SecureBoot   bool
	Virtualized  bool
	SerialNumber string
	// List of gpus short name
	GPUs map[string]interface{}
}

func (r RegistrationInfo) WithGPU(short string) RegistrationInfo {
	if r.GPUs == nil {
		r.GPUs = make(map[string]interface{})
	}

	r.GPUs[short] = nil
	return r
}

func (r RegistrationInfo) WithCapacity(v gridtypes.Capacity) RegistrationInfo {
	r.Capacity = v
	return r
}

func (r RegistrationInfo) WithLocation(v geoip.Location) RegistrationInfo {
	r.Location = v
	return r
}

func (r RegistrationInfo) WithSecureBoot(v bool) RegistrationInfo {
	r.SecureBoot = v
	return r
}

func (r RegistrationInfo) WithVirtualized(v bool) RegistrationInfo {
	r.Virtualized = v
	return r
}

func (r RegistrationInfo) WithSerialNumber(v string) RegistrationInfo {
	r.SerialNumber = v
	return r
}

func (r *Registrar) registration(ctx context.Context, cl zbus.Client, env environment.Environment, info RegistrationInfo) (nodeID, twinID uint64, err error) {
	// we need to collect all node information here
	// - we already have capacity
	// - we get the location (will not change after initial registration)
	loc, err := geoip.Fetch()
	if err != nil {
		return 0, 0, errors.Wrap(err, "fetch location")
	}

	log.Debug().
		Uint64("cru", info.Capacity.CRU).
		Uint64("mru", uint64(info.Capacity.MRU)).
		Uint64("sru", uint64(info.Capacity.SRU)).
		Uint64("hru", uint64(info.Capacity.HRU)).
		Msg("node capacity")

	info = info.WithLocation(loc)

	nodeID, twinID, err = registerNode(ctx, env, cl, info)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to register node")
	}

	return nodeID, twinID, nil
}

func retryNotify(err error, d time.Duration) {
	log.Warn().Err(err).Str("sleep", d.String()).Msg("registration failed")
}

func registerNode(
	ctx context.Context,
	env environment.Environment,
	cl zbus.Client,
	info RegistrationInfo,
) (nodeID, twinID uint64, err error) {
	var (
		mgr              = zos4Stubs.NewIdentityManagerStub(cl)
		netMgr           = stubs.NewNetworkerLightStub(cl)
		registrarGateway = zos4Stubs.NewRegistrarGatewayStub(cl)
	)

	infs, err := netMgr.Interfaces(ctx, "zos", "")
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to get zos bridge information")
	}

	interfaces := []client.Interface{{
		Name: infs.Interfaces["zos"].Name,
		Mac:  infs.Interfaces["zos"].Mac,
		IPs: func() string {
			ips := make([]string, 0)
			for _, ip := range infs.Interfaces["zos"].IPs {
				ips = append(ips, ip.IP.String())
			}
			return strings.Join(ips, "/")
		}(),
	}}

	resources := client.Resources{
		HRU: uint64(info.Capacity.HRU),
		SRU: uint64(info.Capacity.SRU),
		CRU: info.Capacity.CRU,
		MRU: uint64(info.Capacity.MRU),
	}

	location := client.Location{
		Longitude: fmt.Sprint(info.Location.Longitude),
		Latitude:  fmt.Sprint(info.Location.Latitude),
		Country:   info.Location.Country,
		City:      info.Location.City,
	}

	log.Info().Str("id", mgr.NodeID(ctx).Identity()).Msg("start registration of the node on zos4 registrar")

	account, err := registrarGateway.EnsureAccount(ctx, env.RelayURL, "")
	if err != nil {
		log.Info().Msg("failed to EnsureAccount")
		return 0, 0, errors.Wrap(err, "failed to ensure account")
	}
	twinID = account.TwinID

	serial := info.SerialNumber

	real := client.Node{
		FarmID:       uint64(env.FarmID),
		TwinID:       twinID,
		Resources:    resources,
		Location:     location,
		Interfaces:   interfaces,
		SecureBoot:   info.SecureBoot,
		Virtualized:  info.Virtualized,
		SerialNumber: serial,
	}

	node, regErr := registrarGateway.GetNodeByTwinID(ctx, twinID)
	nodeID = node.NodeID
	if regErr != nil {
		if strings.Contains(regErr.Error(), client.ErrorNodeNotFround.Error()) {
			nodeID, err = registrarGateway.CreateNode(ctx, real)
			if err != nil {
				return 0, 0, errors.Wrap(err, "failed to create node on registrar")
			}
		} else {
			return 0, 0, errors.Wrapf(regErr, "failed to get node information for twin id: %d", twinID)
		}
	}

	// node exists
	var onRegistrar client.Node
	onRegistrar, err = registrarGateway.GetNode(ctx, nodeID)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "failed to get node with id: %d", nodeID)
	}

	// ignore virt-what value if the node is marked as real on the registrar
	if !onRegistrar.Virtualized {
		real.Virtualized = false
	}

	real.NodeID = nodeID

	// node exists. we validate everything is good
	// otherwise we update the node
	log.Debug().Uint64("node", nodeID).Msg("node already found on registrar")

	if !reflect.DeepEqual(real, onRegistrar) {
		log.Debug().Msgf("node data have changed, issuing an update node real: %+v\nonRegistrar: %+v", real, onRegistrar)
		err := registrarGateway.UpdateNode(ctx, real)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "failed to update node data with id: %d", nodeID)
		}
	}

	return nodeID, twinID, err
}
