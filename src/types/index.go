package types

type ClientOpts struct {
	LogLevel   string                 `json:"logLevel"`
	ServerHost string                 `json:"serverHost"`
	ServerPort int                    `json:"serverPort"`
	Tunnels    map[string]*TunnelOpts `json:"tunnels"`
}

type TunnelOpts struct {
	Name       string `json:"string"`
	Type       string `json:"type"`
	LocalPort  int    `json:"localPort"`
	RemotePort int    `json:"remotePort"`
	Subdomain  string `json:"subdomain,omitempty"`
	Status     string
}

func GetTypeByNo(ptype uint8) string {
	if ptype == 0x1 {
		return "tcp"
	} else {
		return "web"
	}
}

func (tunopts *TunnelOpts) GetTypeNo() uint8 {
	if tunopts.Type == "tcp" {
		return 0x1
	} else {
		return 0x2
	}
}

type ServerOpts struct {
	LogLevel   string `json:"logLevel"`
	ServerHost string `json:"serverHost"`
	ServerPort int    `json:"serverPort"`
	HttpAddr   int    `json:"httpAddr"`
	HttpsAddr  int    `json:"httpsAddr"`
	TlsCA      string `json:"tlsCA"`
	TlsCrt     string `json:"tlsCrt"`
	TlsKey     string `json:"tlsKey"`
}

type StatusMsg struct {
	Status  uint8
	Message string
}
