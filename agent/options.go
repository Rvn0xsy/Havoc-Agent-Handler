package main

/*
{
'Arch': 'x64',
'Format': 'Windows-Agent',
'Listener': {
'BehindRedir': False,
'Cert': {'Cert': '', 'Key': ''},
'Headers': [''],
'HostBind': '10.100.12.40',
'HostHeader': '',
'HostRotation': 'round-robin',
'Hosts': ['10.100.12.40'],
'KillDate': 0, 'Methode': '',
'Name': 'Hello',
'PortBind': '8090',
'PortConn': '',
'Proxy': {'Enabled': False, 'Host': '', 'Password': '', 'Port': '', 'Type': '', 'Username': ''},
'Response': {'Headers': ['']},
'Secure': False,
'Uris': [''],
'UserAgent': 'Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36',
'WorkingHours': ''}}
*/

type Config struct {
	Arch     string `json:"Arch"`
	Format   string `json:"Format"`
	Listener struct {
		BehindRedir  bool     `json:"BehindRedir"`
		Cert         Cert     `json:"Cert"`
		Headers      []string `json:"Headers"`
		HostBind     string   `json:"HostBind"`
		HostHeader   string   `json:"HostHeader"`
		HostRotation string   `json:"HostRotation"`
		Hosts        []string `json:"Hosts"`
		KillDate     int      `json:"KillDate"`
		Methode      string   `json:"Methode"`
		Name         string   `json:"Name"`
		PortBind     string   `json:"PortBind"`
		PortConn     string   `json:"PortConn"`
		Proxy        Proxy    `json:"Proxy"`
		Response     Response `json:"Response"`
		Secure       bool     `json:"Secure"`
		Uris         []string `json:"Uris"`
		UserAgent    string   `json:"UserAgent"`
		WorkingHours string   `json:"WorkingHours"`
	} `json:"Listener"`
}

type Cert struct {
	Cert string `json:"Cert"`
	Key  string `json:"Key"`
}

type Proxy struct {
	Enabled  bool   `json:"Enabled"`
	Host     string `json:"Host"`
	Password string `json:"Password"`
	Port     string `json:"Port"`
	Type     string `json:"Type"`
	Username string `json:"Username"`
}

type Response struct {
	Headers []string `json:"Headers"`
}

var (
	OptionsConfigString = `OPTIONS_STRING`
	OptionsConfig       = Config{}
)
