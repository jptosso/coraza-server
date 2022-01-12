package spoa

import (
	"fmt"

	spoe "github.com/criteo/haproxy-spoe-go"
	"github.com/jptosso/coraza-waf/v2"
	"github.com/sirupsen/logrus"
)

func (s *SPOA) processResponse(msg spoe.Message) ([]spoe.Action, error) {

	phase := 0
	status := 0
	httpver := ""
	var tx *coraza.Transaction
	// status res.ver res.hdrs res.body_size res.body
	argnames := []string{"Transaction ID", "Response Code", "Response Headers", "Response Body Size", "Response Body"}
	logrus.Debug("Attempting to process response")
	for msg.Args.Next() {
		arg := msg.Args.Arg
		value := ""
		if phase == 0 || phase == 2 || phase == 3 || phase == 4 {
			var ok bool
			value, ok = arg.Value.(string)
			if !ok && (phase == 0 || phase == 2) {
				return nil, fmt.Errorf("invalid argument for %s, string expected, got %v", argnames[phase], arg.Value)
			}
		}
		logrus.Debugf("Phase %d", phase)
		switch phase {
		case 0:
			// TX UNIQUE ID
			tx = s.txcache.Get(value)
			if tx == nil {
				return nil, fmt.Errorf("attempting to process expired transaction")
			}
		case 1:
			// Status
			val, ok := arg.Value.(int)
			if !ok {
				return nil, fmt.Errorf("invalid response code, got %s", arg.Value)
			}
			logrus.Debugf("Got response status %d", val)
			status = val
		case 2:
			// HTTP VERSION
			logrus.Debugf("Got HTTP Version %s", value)
			httpver = value
		case 3:
			// RESPONSE HEADERS
			h, err := readHeaders(value)
			if err != nil {
				return nil, err
			}
			for k, vv := range h {
				for _, v := range vv {
					tx.AddResponseHeader(k, v)
				}
			}
			if it := tx.ProcessResponseHeaders(status, httpver); it != nil {
				return spoeFail(true), nil
			}
		case 4:
			// RESPONSE BODY SIZE
			//bsize = value
		case 5:
			// RESPONSE BODY
			if body, ok := arg.Value.([]byte); ok {
				tx.ResponseBodyBuffer.Write(body)
			}
			tx.ProcessResponseBody()
		default:
			logrus.Error("Unexpected message")
		}
		phase++
	}
	if it := tx.Interruption; it != nil {
		// Transaction disrupted
		return spoeFail(true), nil
	}

	//Expire will also run tx.ProcessLogging
	if err := s.txcache.Expire(tx.ID); err != nil {
		// what to do here?
		logrus.Error("Failed to expire transaction")
	}

	logrus.Debug("Transaction was not disrupted")
	// Just pass
	return spoeFail(false), nil
}
