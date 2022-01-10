package spoa

import (
	"fmt"
	"net"

	spoe "github.com/criteo/haproxy-spoe-go"
	"github.com/jptosso/coraza-waf/v2"
	"github.com/jptosso/coraza-waf/v2/types/variables"
	"github.com/sirupsen/logrus"
)

func (s *SPOA) processRequest(msg spoe.Message) ([]spoe.Action, error) {

	phase := 0
	var method, path, query, httpv string
	var tx *coraza.Transaction
	argnames := []string{"Transaction ID", "Request IP", "Method", "Path", "Query", "HTTP Version",
		"Request Headers", "Request Body Size", "Request Body"}
	for msg.Args.Next() {
		arg := msg.Args.Arg
		value := ""
		if phase == 0 || (phase >= 2 && phase < 7) || phase == 8 {
			ok := true
			value, ok = arg.Value.(string)
			if !ok && (phase == 0 || phase == 2) {
				return nil, fmt.Errorf("invalid argument for %s, string expected, got %v", argnames[phase], arg.Value)
			}
		}
		logrus.Infof("Phase %d", phase)
		switch phase {
		case 0:
			// TX UNIQUE ID
			tx = s.waf.NewTransaction()
			tx.ID = value
			tx.GetCollection(variables.UniqueID).Set("", []string{tx.ID})
			if err := s.txcache.Store(tx); err != nil {
				return nil, err
			}
		case 1:
			// REQUEST IP
			val, ok := arg.Value.(net.IP)
			if !ok {
				return nil, fmt.Errorf("invalid ip address")
			}
			logrus.Debugf("Got request ip %s", val.String())
			tx.ProcessConnection(val.String(), 0, "", 0)
		case 2:
			// METHOD
			method = value
		case 3:
			// PATH
			path = value
		case 4:
			//QUERY
			query = value
			//tx.ProcessConnection()
		case 5:
			// HTTP VERSION
			httpv = value
			tx.ProcessURI(path+"?"+query, method, httpv)
		case 6:
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
			if it := tx.ProcessRequestHeaders(); it != nil {
				return spoeFail(true), nil
			}
		case 7:
			// REQUEST BODY SIZE
			//bsize = value
		case 8:
			// REQUEST BODY
			tx.ResponseBodyBuffer.Write([]byte(value))
			tx.ProcessRequestBody()
		default:
			logrus.Error("Unexpected message")
		}
		phase++
	}
	if it := tx.Interruption; it != nil {
		// Transaction disrupted
		return spoeFail(true), nil
	}

	logrus.Debug("Transaction was not disrupted")
	// Just pass
	return spoeFail(false), nil
}
