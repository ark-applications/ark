package proxy

import (
	"os"
	"strconv"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/dkimot/ark/arkd/internal/config"
	"github.com/oklog/ulid/v2"
)

var RootInternalDomain = ".internal"

type Proxy interface {
  RegisterApp(id, name, domainName, port string) error
  DelistApp(id string) error
}

func New(cfg config.Config) Proxy {
  p := &proxy{
    registeredApps: make(map[string]AppDefinition),
  }

  p.RegisterApp(ulid.Make().String(), "arkd-worker-api", "arkd-worker-api", strconv.Itoa(cfg.ApiPort))

  return p
}

type proxy struct {
  mtx            sync.Mutex
  registeredApps map[string]AppDefinition
}

func (p *proxy) RegisterApp(id, name, domainName, port string) error {
  p.mtx.Lock()
  defer p.mtx.Unlock()

  p.registeredApps[id] = AppDefinition{
    ID: id,
    Name: name,
    DomainName: domainName + RootInternalDomain,
    Port: port,
  }

  return p.writeConfig()
}

func (p *proxy) DelistApp(id string) error {
  p.mtx.Lock()
  defer p.mtx.Unlock()

  delete(p.registeredApps, id)

  return p.writeConfig()
}

func (p *proxy) writeConfig() error {
  cfgApps := map[string]ProxyConfigApp{}

  for _, app := range p.registeredApps {
    cfgApps[app.Name] = newAppConfig(app)
  }

  cfg := ProxyConfig{
    ListenPort: ProxyListenPort,
    ListenPortTls: ProxyListenPortTls,
    Apps: cfgApps,
  }

  buf, err := toml.Marshal(cfg)
  if err != nil {
    return err
  }
  if err := os.WriteFile("./rpxy-config/rpxy.toml", buf, 0600); err != nil {
    return nil
  }

  return nil

}

type AppDefinition struct {
  ID         string
  Name       string
  DomainName string
  Port       string
}
