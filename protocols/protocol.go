package protocols

import (
	"fmt"

	"github.com/jptosso/coraza-server/config"
	"github.com/jptosso/coraza-server/protocols/spoa"
	"github.com/jptosso/coraza-waf/v2"
)

type Protocol interface {
	Init(waf *coraza.Waf, cfg config.Agent) error
	Start() error
}

func GetProtocol(name string) (Protocol, error) {
	switch name {
	case "spoa":
		return &spoa.SPOA{}, nil
	default:
		return nil, fmt.Errorf("invalid protocol %s", name)
	}
}
