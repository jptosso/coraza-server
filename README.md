Coraza Server is the most ambitious implementation of Coraza WAF, it's designed to integrate with systems written in different languages, like C, using multiple protocols like SPOA, REST and GRPC.

## Installing

To install coraza library you are required to have a C compiler, libinjection and pcre installed, see [https://coraza.io/docs/tutorials/dependencies/](https://coraza.io/docs/tutorials/dependencies/)

If you cannot install these dependencies you may 
```sh
go install github.com/jptosso/coraza-server/cmd/coraza-server@master
```

## Using as a Container


## Using in K8


## Configuration

**Configuration is not stable yet.**

## Protocol development status

### SPOA

- **API Stability:** Unstable
- **Code Stability:** Unstable
- **Documentation:** Not available yet

### REST

- **API Stability:** Not designed yet
- **Code Stability:** Not written
- **Documentation:** Not available yet

### GRPC

- **API Stability:** Under development
- **Code Stability:** Not written
- **Documentation:** Not available yet

## Installing plugins

To install Coraza plugins you must copy the content from ```cmd/coraza-server/main.go``` and add the dependencies named with _, for example:

```go
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

    // Plugins
    _ "github.com/path/to-plugin"
)
```

## References

* [Set up using haproxy](#)
* [Setting up a k8 cluster with haproxy](#)
* [REST API References](#)
* [GRPC Proto and documentation](#)

## TODO

- [ ] Add workers limit to SPOP
- [ ] Document SPOP
- [ ] Create REST protocol
- [ ] Create GRPC protocol
- [ ] Normalize settings
- [ ] Regression tests
- [ ] Replace SPOA library with a custom one