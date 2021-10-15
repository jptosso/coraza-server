package spoa

import (
	"fmt"

	spoe "github.com/criteo/haproxy-spoe-go"
	"github.com/jptosso/coraza-waf"
)

func (s *SPOA) processRequest(msg spoe.Message) ([]spoe.Action, error) {

	phase := 0
	var method, path, query, httpv string
	var tx *coraza.Transaction
	argnames := []string{"WAF Id", "Transaction ID", "Request IP", "Method", "Path", "Query", "HTTP Version",
		"Request Headers", "Request Body Size", "Request Body"}
	for msg.Args.Next() {
		arg := msg.Args.Arg
		value, ok := arg.Value.(string)
		if !ok && phase != 7 {
			return nil, fmt.Errorf("invalid argument for %s, string expected", argnames[phase])
		}
		switch phase {
		case 0:
			// TX UNIQUE ID
			tx = s.waf.NewTransaction()
			tx.Id = value
			tx.GetCollection(coraza.VARIABLE_UNIQUE_ID).Set("", []string{tx.Id})
			if err := s.txcache.Store(tx); err != nil {
				return nil, err
			}
		case 1:
			// REQUEST IP
			tx.ProcessConnection(value, 0, "", 0)
		case 2:
			// METHOD
			method = value
		case 4:
			// PATH
			path = value
		case 5:
			//QUERY
			query = value
			//tx.ProcessConnection()
		case 6:
			// HTTP VERSION
			httpv = value
			tx.ProcessUri(path+query, method, httpv)
		case 7:
			// RESQUEST HEADERS
			h, err := readHeaders(value)
			if err != nil {
				return nil, err
			}
			for k, vv := range h {
				for _, v := range vv {
					tx.AddRequestHeader(k, v)
				}
			}
			tx.ProcessRequestHeaders()
		case 8:
			// REQUEST BODY SIZE
			//bsize = value
		case 9:
			// REQUEST BODY
			tx.ResponseBodyBuffer.Write([]byte(value))
			tx.ProcessRequestBody()
		default:
			fmt.Println("Unexpected message")
		}
		phase++
	}
	if it := tx.Interruption; it != nil {
		// Transaction disrupted
		return txToAction(tx), nil
	}

	// Just pass
	return []spoe.Action{
		spoe.ActionSetVar{
			Name:  "action",
			Scope: spoe.VarScopeSession,
			Value: "pass",
		},
	}, nil
}

func txToAction(tx *coraza.Transaction) []spoe.Action {
	if tx.Interrupted() {
		return []spoe.Action{
			spoe.ActionSetVar{
				Name:  "action",
				Scope: spoe.VarScopeSession,
				Value: tx.Interruption.Action,
			},
		}
	} else {
		return []spoe.Action{
			spoe.ActionSetVar{
				Name:  "action",
				Scope: spoe.VarScopeSession,
				Value: "pass",
			},
		}
	}
}
