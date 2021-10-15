package main

import (
	"flag"
	"os"
	"sync"

	"github.com/jptosso/coraza-server/config"
	"github.com/jptosso/coraza-server/protocols"
	"github.com/jptosso/coraza-waf"
	"github.com/jptosso/coraza-waf/seclang"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func main() {
	f := flag.String("f", "", "Absolute path to configuration file (.yaml)")
	flag.Parse()
	cfg, err := readConfig(*f)
	if err != nil {
		logrus.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	for _, a := range cfg.Agents {
		proto, err := protocols.GetProtocol(a.Protocol)
		if err != nil {
			logrus.Fatal(err)
		}
		wg.Add(1)
		logrus.Info("Initializing waf")
		waf := coraza.NewWaf()
		parser, _ := seclang.NewParser(waf)
		if len(a.Include) == 0 {
			logrus.Warn("No rules detected for agent")
		}
		for _, file := range a.Include {
			if err := parser.FromFile(file); err != nil {
				logrus.Fatal(err)
			}
		}
		logrus.Infof("Initializing protocol %s", a.Protocol)
		proto.Init(waf, a)
		logrus.Infof("Starting protocol %s on %s", a.Protocol, a.Bind)
		go func() {
			defer wg.Done()
			if err := proto.Start(); err != nil {
				logrus.Fatal(err)
			}
		}()
	}
	wg.Wait()
	logrus.Info("Coraza server finished.")
}

func readConfig(path string) (config.Config, error) {
	f, err := os.ReadFile(path)
	var cfg config.Config
	if err != nil {
		return config.Config{}, err
	}
	err = yaml.Unmarshal(f, &cfg)
	return cfg, err
}
