package internal

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

type Network string

var (
	MainNetwork Network = "production-v4"
	TestNetwork Network = "testing-v4"
	QANetwork   Network = "qa-v4"
)

type Params struct {
	Interval time.Duration
	QAUrl    string
	TestUrl  string
	MainUrl  string
}

type Worker struct {
	src string
	dst string

	interval time.Duration
	clients  map[Network]client.RegistrarClient
}

// NewWorker creates a new instance of the worker
func NewWorker(src string, dst string, params Params) (*Worker, error) {
	// we need to recalculate the path of the symlink here because of the following
	// - assume we run the tool like `updater -d dst -s src`
	// - it's then gonna build the links as above.
	// - then it will crease dst/zos:testing-3:latest.flist that points to dst/zos:<v>.flist
	// and that is wrong because now the link points to a wrong path. it instead need to be ../dst/<file>
	// so recalculating here
	// we need to find a abs path from dst to src.
	// so this goes as this
	// - we make sure that src and dst are always abs
	// this later will allow us to calculate relative path from dst to src

	src, err := filepath.Abs(src)
	if err != nil {
		return nil, fmt.Errorf("failed to get src as abs path: %w", err)
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to get dst as abs path: %w", err)
	}

	log.Info().Str("src", src).Str("dst", dst).Msg("paths")

	clients := map[Network]client.RegistrarClient{}

	if params.QAUrl != "" {
		url, err := url.JoinPath(params.QAUrl, "v1")
		if err != nil {
			return nil, err
		}

		cli, err := client.NewRegistrarClient(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new registrar client for qa net")
		}
		clients[QANetwork] = cli
	}

	if params.TestUrl != "" {
		url, err := url.JoinPath(params.TestUrl, "v1")
		if err != nil {
			return nil, err
		}

		cli, err := client.NewRegistrarClient(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new registrar client for test net")
		}
		clients[TestNetwork] = cli
	}

	if params.MainUrl != "" {
		url, err := url.JoinPath(params.MainUrl, "v1")
		if err != nil {
			return nil, err
		}

		cli, err := client.NewRegistrarClient(url)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new registrar client for main net")
		}
		clients[MainNetwork] = cli
	}

	return &Worker{
		src:      src,
		dst:      dst,
		clients:  clients,
		interval: params.Interval,
	}, nil
}

// checkNetwork to check if a network is valid against main, qa, test
func checkNetwork(network Network) error {
	if network != MainNetwork && network != QANetwork && network != TestNetwork {
		return fmt.Errorf("invalid network")
	}

	return nil
}

// updateZosVersion updates the latest zos flist for a specific network with the updated zos version
func (w *Worker) updateZosVersion(network Network, regClient client.RegistrarClient) error {
	if err := checkNetwork(network); err != nil {
		return err
	}

	regVersion, err := regClient.GetZosVersion()
	if err != nil {
		return err
	}

	// var regVersion RegistrarVersion
	// err = json.Unmarshal([]byte(currentZosVersion), &regVersion)
	// if err != nil {
	// 	log.Debug().Err(err).Msg("failed to unmarshal chain version")
	// 	// shouldn't fail for env that still not updated version format
	// 	return nil
	// }

	log.Debug().Msgf("getting substrate version %v for network %v", regVersion.Version, network)

	// now we need to find how dst is relative to src
	path, err := filepath.Rel(w.dst, w.src)
	if err != nil {
		return fmt.Errorf("failed to get dst relative path to src: %w", err)
	}

	// zos
	zosCurrent := fmt.Sprintf("%v/.tag-%v", w.src, regVersion.Version)
	zosLatest := fmt.Sprintf("%v/%v", w.dst, network)

	// the link is like zosCurrent but it has the path relative from the symlink
	// point of view (so relative to the symlink, how to reach zosCurrent)
	// hence the link is instead used in all calls to symlink
	zosLink := fmt.Sprintf("%v/.tag-%v", path, regVersion.Version)

	// update links zos
	return w.updateLink(zosCurrent, zosLatest, zosLink)
}

func (w *Worker) updateLink(current string, latest string, link string) error {
	// check if current exists
	if _, err := os.Lstat(current); err != nil {
		return err
	}

	// check if symlink exists
	dst, err := os.Readlink(latest)

	// if no symlink, then create it
	if os.IsNotExist(err) {
		log.Info().Str("from", latest).Str("to", current).Msg("linking")
		return os.Symlink(link, latest)
	} else if err != nil {
		return err
	}

	// check if symlink is valid and exists
	if filepath.Base(dst) == filepath.Base(current) {
		log.Debug().Msgf("symlink %v to %v already exists", current, latest)
		return nil
	}

	// remove symlink if it is not valid and exists
	if err := os.Remove(latest); err != nil {
		return err
	}

	log.Info().Str("from", latest).Str("to", current).Msg("linking")
	return os.Symlink(link, latest)
}

// UpdateWithInterval updates the latest zos flist for a specific network with the updated zos version
// with a specific interval between each update
func (w *Worker) UpdateWithInterval(ctx context.Context) {
	ticker := time.NewTicker(w.interval)

	for {
		for network, regClient := range w.clients {
			log.Debug().Msgf("updating zos version for %v", network)

			exp := backoff.NewExponentialBackOff()
			exp.MaxInterval = 2 * time.Second
			exp.MaxElapsedTime = 10 * time.Second
			err := backoff.Retry(func() error {
				err := w.updateZosVersion(network, regClient)
				if err != nil {
					log.Error().Err(err).Msg("update failure. retrying")
				}
				return err
			}, exp)
			if err != nil {
				log.Error().Err(err).Msg("update zos failed with error")
			}
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}
