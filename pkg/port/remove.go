package port

import (
	"fmt"

	"github.com/neaas/neslink"
)

func PortPairRemove(nsp neslink.NsProvider, port Port) error {
	if err := neslink.Do(
		nsp,
		neslink.LADelete(neslink.LPIndex(port.Attrs().Index)),
	); err != nil {
		return fmt.Errorf("could not delete port pair %s: %w", port.Attrs().Name, err)
	}
	return nil
}
