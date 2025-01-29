package provision

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"github.com/threefoldtech/zos4/pkg/stubs"
)

type registrarTwins struct {
	registrarGateway *stubs.RegistrarGatewayStub
	mem              *lru.Cache
}

// NewRegistrarTwins creates a users db that implements the provision.Users interface.
func NewRegistrarTwins(registrarGateway *stubs.RegistrarGatewayStub) (Twins, error) {
	cache, err := lru.New(1024)
	if err != nil {
		return nil, err
	}

	return &registrarTwins{
		registrarGateway: registrarGateway,
		mem:              cache,
	}, nil
}

// GetKey gets twins public key
func (s *registrarTwins) GetKey(id uint32) ([]byte, error) {
	if value, ok := s.mem.Get(id); ok {
		return value.([]byte), nil
	}
	user, err := s.registrarGateway.GetTwin(context.Background(), uint64(id))
	if err != nil {
		return nil, errors.Wrapf(err, "could not get user with id '%d'", id)
	}

	key, err := base64.StdEncoding.DecodeString(user.PublicKey)
	if err != nil {
		return nil, errors.Wrapf(err, "could decode public key for user with id '%d'", id)
	}

	s.mem.Add(id, key)
	return key, nil
}

type registrarAdmins struct {
	twin uint32
	pk   ed25519.PublicKey
}

// NewRegistrarAdmins creates a twins db that implements the provision.Users interface.
// but it also make sure the user is an admin
func NewRegistrarAdmins(registrarGateway *stubs.RegistrarGatewayStub, farmID uint64) (Twins, error) {
	farm, err := registrarGateway.GetFarm(context.Background(), farmID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get farm")
	}

	twin, err := registrarGateway.GetTwin(context.Background(), farm.TwinID)
	if err != nil {
		return nil, err
	}

	key, err := base64.StdEncoding.DecodeString(twin.PublicKey)
	if err != nil {
		return nil, errors.Wrapf(err, "could decode public key for twin with farm id '%d'", farmID)
	}

	return &registrarAdmins{
		twin: uint32(farm.TwinID),
		pk:   key,
	}, nil
}

// GetKey gets twin public key if twin is valid admin
func (s *registrarAdmins) GetKey(id uint32) ([]byte, error) {
	if id != s.twin {
		return nil, fmt.Errorf("twin with id '%d' is not an admin", id)
	}

	return []byte(s.pk), nil
}
