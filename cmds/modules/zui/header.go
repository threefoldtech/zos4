package zui

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/gizak/termui/v3/widgets"
	"github.com/pkg/errors"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/zbus"
	registrar "github.com/threefoldtech/zos4/pkg/registrar_light"
	zos4Stubs "github.com/threefoldtech/zos4/pkg/stubs"
	"github.com/threefoldtech/zosbase/pkg/app"
	"github.com/threefoldtech/zosbase/pkg/environment"
	"github.com/threefoldtech/zosbase/pkg/stubs"
)

func green(s string) string {
	return fmt.Sprintf("[%s](fg:green)", s)
}

func red(s string) string {
	return fmt.Sprintf("[%s](fg:red)", s)
}

func isInProgressError(err error) bool {
	return strings.Contains(err.Error(), registrar.ErrInProgress.Error())
}

// func headerRenderer(c zbus.Client, h *widgets.Paragraph, r *Flag) error {
func headerRenderer(ctx context.Context, c zbus.Client, h *widgets.Paragraph, r *signalFlag) error {
	env, err := environment.Get()
	if err != nil {
		return err
	}

	identity := zos4Stubs.NewIdentityManagerStub(c)
	registrar := zos4Stubs.NewRegistrarStub(c)

	h.Text = "\n    Fetching realtime node information... please wait."

	s := "          Welcome to [Zero-OS](fg:yellow), [ThreeFold](fg:blue) Autonomous Operating System\n" +
		"\n" +
		" This is node %s (farmer %s)\n" +
		" running Zero-OS version [%s](fg:blue) (mode [%s](fg:cyan))\n" +
		" kernel: %s\n" +
		" cache disk: %s"

	host := stubs.NewVersionMonitorStub(c)
	ch, err := host.Version(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start update stream for version")
	}

	go func() {
		registrarLable := "registrar"
		zui := stubs.NewZUIStub(c)

		// empty out zui errors for registrar
		if zuiErr := zui.PushErrors(ctx, registrarLable, []string{}); zuiErr != nil {
			log.Info().Err(zuiErr).Send()
		}

		farmID := identity.FarmID(ctx)
		for version := range ch {
			var name string
			var nodeID string
			var farm string
			if name, err = identity.Farm(ctx); err != nil {
				farm = red(fmt.Sprintf("%d: %s", farmID, err.Error()))
			} else {
				farm = green(fmt.Sprintf("%d: %s", farmID, name))
			}

			if node, err := registrar.NodeID(ctx); err != nil {
				if isInProgressError(err) {
					nodeID = green(err.Error())
				} else {
					nodeID = red(fmt.Sprintf("%d (unregistered)", node))
					if zuiErr := zui.PushErrors(ctx, registrarLable, []string{err.Error()}); zuiErr != nil {
						log.Info().Err(zuiErr).Send()
					}
				}
			} else {
				nodeID = green(fmt.Sprint(node))
			}

			cache := green("OK")
			if app.CheckFlag(app.LimitedCache) {
				cache = red("no ssd disks detected, running on hdd-only mode")
			} else if app.CheckFlag(app.ReadonlyCache) {
				cache = red("cache is read-only")
			}

			var utsname syscall.Utsname
			var uname string
			if err := syscall.Uname(&utsname); err != nil {
				uname = red(err.Error())
			} else {
				uname = green(string(unsafe.Slice((*byte)(unsafe.Pointer(&utsname.Release)), len(utsname.Release))))
			}

			h.Text = fmt.Sprintf(s, nodeID, farm, version.String(), env.RunningMode.String(), uname, cache)
			r.Signal()
		}
	}()

	return nil
}
