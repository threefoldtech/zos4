package power

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/host"
	"github.com/threefoldtech/substrate-client"
	"github.com/threefoldtech/zos/pkg/utils"
)

const (
	reportUptimeEvery = 40 * time.Minute
)

type Uptime struct {
	// Mark is set to done after the first uptime is sent
	Mark utils.Mark

	id  substrate.Identity
	sub substrate.Manager
	m   sync.Mutex
}

func NewUptime(sub substrate.Manager, id substrate.Identity) (*Uptime, error) {
	return &Uptime{
		id:   id,
		sub:  sub,
		Mark: utils.NewMark(),
	}, nil
}

func (u *Uptime) SendNow() (types.Hash, error) {
	uptime, err := host.Uptime()
	if err != nil {
		return types.Hash{}, errors.Wrap(err, "failed to get uptime")
	}
	return u.send(uptime)
}

func (u *Uptime) send(uptime uint64) (types.Hash, error) {
	// the mutex is to avoid race when SendNow is called
	// while the times reporting is working
	u.m.Lock()
	defer u.m.Unlock()

	sub, err := u.sub.Substrate()
	if err != nil {
		return types.Hash{}, err
	}
	defer sub.Close()
	return sub.UpdateNodeUptime(u.id, uptime)
}

func (u *Uptime) uptime(ctx context.Context) error {
	for {
		uptime, err := host.Uptime()
		if err != nil {
			return errors.Wrap(err, "failed to get uptime")
		}
		log.Debug().Msg("updating node uptime")
		hash, err := u.send(uptime)
		if err != nil {
			return errors.Wrap(err, "failed to report uptime")
		}

		u.Mark.Signal()

		log.Info().Str("hash", hash.Hex()).Msg("node uptime hash")

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(reportUptimeEvery):
			continue
		}
	}
}

// start uptime reporting. returns a channel that is closed immediately after
// the first uptime is reported.
func (u *Uptime) Start(ctx context.Context) {
	// uptime update
	defer log.Info().Msg("uptime reporting exited permanently")
	safeUptime := func(ctx context.Context) (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("uptime reporting has panicked: %+v\n%s", p, string(debug.Stack()))
			}
		}()

		err = u.uptime(ctx)
		return err
	}

	for {
		err := safeUptime(ctx)
		if errors.Is(err, context.Canceled) {
			log.Info().Msg("stop uptime reporting. context cancelled")
			return
		} else if err != nil {
			log.Error().Err(err).Msg("sending uptime failed")
		} else {
			// context was cancelled
			return
		}
		// even there is no error we try again until ctx is cancelled
		<-time.After(10 * time.Second)
	}
}