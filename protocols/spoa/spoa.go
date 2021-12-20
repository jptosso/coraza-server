package spoa

import (
	"fmt"
	"net/http"
	"strings"

	spoe "github.com/criteo/haproxy-spoe-go"
	"github.com/jptosso/coraza-server/cache"
	"github.com/jptosso/coraza-server/config"
	"github.com/jptosso/coraza-waf/v2"
	log "github.com/sirupsen/logrus"
)

type SPOA struct {
	cfg     config.Agent
	waf     *coraza.Waf
	txcache *cache.TxCache
}

func (s *SPOA) Init(waf *coraza.Waf, cfg config.Agent) error {
	s.cfg = cfg
	s.waf = waf
	s.txcache = cache.NewTxCache(cfg.TransactionTtl, cfg.TxActiveLimit)
	if cfg.Workers > 0 {
		log.Warn("Workers are not supported by SPOP protocol, it will be ignored")
	}
	return nil
}

func (s *SPOA) Start() error {
	log.Info("Registering SPOP agent")
	agent := spoe.New(func(messages *spoe.MessageIterator) ([]spoe.Action, error) {
		for messages.Next() {
			msg := messages.Message

			if msg.Name == "coraza-req" {
				return s.processRequest(msg)
			}

			if msg.Name == "coraza-res" {
				return s.processResponse(msg)
			}
		}
		return nil, fmt.Errorf("invalid protocol request")
	})
	log.Info("Starting SPOP server")
	if err := agent.ListenAndServe(s.cfg.Bind); err != nil {
		log.Fatal(err)
	}
	return nil
}

func readHeaders(headers string) (http.Header, error) {
	h := http.Header{}
	spl := strings.Split(headers, "\r\n")
	for _, l := range spl {
		if l == "" {
			continue
		}
		spl2 := strings.SplitN(l, ":", 2)
		if len(spl2) != 2 {
			return nil, fmt.Errorf("invalid headers format")
		}
		key := strings.TrimSpace(spl2[0])
		value := strings.TrimSpace(spl2[1])
		h.Add(key, value)
	}
	return h, nil
}

func spoeFail(fail bool) []spoe.Action {
	f := 0
	if fail {
		f = 1
	}
	log.Debugf("Sending %t fail signal", fail)
	return []spoe.Action{
		spoe.ActionSetVar{
			Name:  "fail",
			Scope: spoe.VarScopeTransaction,
			Value: f,
		},
	}
}
