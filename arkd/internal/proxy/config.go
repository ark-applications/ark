package proxy

const ProxyListenPort = 8080
const ProxyListenPortTls = 4443

func newAppConfig(app AppDefinition) ProxyConfigApp {
  return ProxyConfigApp{
    ServerName: app.DomainName,
    ReverseProxy: []ProxyConfigAppReverseProxy{{
      Upstream: []ProxyConfigUpstream{{
        Location: "localhost:" + app.Port,
      }},
    }},
  }
}

type ProxyConfigUpstream struct {
  Location string `toml:"location"`
}

type ProxyConfigAppReverseProxy struct {
  Upstream []ProxyConfigUpstream `toml:"upstream"`
  }

type ProxyConfigApp struct {
  ServerName string `toml:"server_name"`
  ReverseProxy []ProxyConfigAppReverseProxy `toml:"reverse_proxy"`
}

type ProxyConfig struct {
  ListenPort    int `toml:"listen_port"`
  ListenPortTls int `toml:"listen_port_tls"`
  Apps          map[string]ProxyConfigApp `toml:"apps"`
}

