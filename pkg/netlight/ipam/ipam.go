package ipam

import (
	"net"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/allocator"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/disk"
	"github.com/rs/zerolog/log"
)

// AllocateIPv4 allocates a unique IPv4 for the entity defines by the given id (for example container id, or a vm).
// in the network with netID, and NetResource.
func AllocateIPv4(networkID, leaseDir string) (*net.IPNet, error) {
	store, err := disk.New("ndmz", leaseDir)
	if err != nil {
		return nil, err
	}

	defer store.Close()

	r := allocator.Range{
		RangeStart: net.ParseIP("100.127.0.3"),
		RangeEnd:   net.ParseIP("100.127.255.254"),
		Subnet: types.IPNet(net.IPNet{
			IP:   net.ParseIP("100.127.0.0"),
			Mask: net.CIDRMask(16, 32),
		}),
		Gateway: net.ParseIP("100.127.0.1"),
	}

	if err := r.Canonicalize(); err != nil {
		return nil, err
	}

	set := allocator.RangeSet{r}

	// unfortunately, calling the allocator Get() directly will try to allocate
	// a new IP. if the ID/nic already has an ip allocated it will just fail instead of returning
	// the same IP.
	// So we have to check the store ourselves to see if there is already an IP allocated
	// to this container, and if one found, we return it.
	if err := store.Lock(); err != nil {
		return nil, err
	}
	ips := store.GetByID(networkID, "eth0")
	_ = store.Unlock()
	if len(ips) > 0 {
		ip := ips[0]
		rng, err := set.RangeFor(ip)
		if err != nil {
			return nil, err
		}

		return &net.IPNet{IP: ip, Mask: rng.Subnet.Mask}, nil
	}

	alloc := allocator.NewIPAllocator(&set, store, 0)

	ipConfig, err := alloc.Get(networkID, "eth0", nil)
	if err != nil {
		return nil, err
	}
	return &ipConfig.Address, nil
}

func DeAllocateIPv4(networkID, leaseDir string) error {
	store, err := disk.New("ndmz", leaseDir)
	if err != nil {
		return err
	}

	defer store.Close()

	// unfortunately, calling the allocator Get() directly will try to allocate
	// a new IP. if the ID/nic already has an ip allocated it will just fail instead of returning
	// the same IP.
	// So we have to check the store ourselves to see if there is already an IP allocated
	// to this container, and if one found, we return it.
	if err := store.Lock(); err != nil {
		return err
	}

	defer func() {
		if err := store.Unlock(); err != nil {
			log.Error().Err(err).Msg("failed to unlock store while deallocating ipv4")
		}
	}()
	return store.ReleaseByID(networkID, "eth0")
}
