package gwk


type ClientOpts struct {
	LogLevel   string `json:"logLevel"`
	TunnelHost string           `json:"tunnelHost"`
	TunnelAddr int              `json:"tunnelAddr"`
	Tunnels    map[string]TunnelConfig `json:"tunnels"`
}


type TunnelConfig struct {
	Name      string  `json:"string"`
	Protocol   string `json:"protocol"`
	LocalPort  int    `json:"localPort"`
	RemotePort int    `json:"remotePort"`
	Subdomain  string `json:"subdomain,omitempty"`
}


type ServerOpts struct {
	LogLevel   string `json:"logLevel"`
	ServerHost string `json:"serverHost"`
	TunnelAddr int    `json:"tunnelAddr"`
	HttpAddr   int    `json:"httpAddr"`
	HttpsAddr  int    `json:"httpsAddr"`
	TlsCA     string `json:"tlsCA"`
	TlsCrt     string `json:"tlsCrt"`
	TlsKey     string `json:"tlsKey"`
}