package napcat

import (
	"fmt"
)

type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

type APIResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type AuthRequest struct {
	Token string `json:"token"`
}

type AuthResponse struct {
	APIResponse
	Data struct {
		Credential string `json:"Credential"`
	} `json:"data"`
}

type SetOB11ConfigRequest struct {
	Config string `json:"config"`
}

type CheckLoginResponse struct {
	APIResponse
	Data struct {
		IsLogin   bool   `json:"isLogin"`
		QRCodeURL string `json:"qrcodeurl"`
	} `json:"data"`
}

type GetLoginInfoResponse struct {
	APIResponse
	Data struct {
		Uin    string `json:"uin"`
		Nick   string `json:"nick"`
		Online bool   `json:"online"`
	} `json:"data"`
}

type WebUIConfig struct {
	Host      string `json:"host"`
	Port      uint32 `json:"port"`
	Prefix    string `json:"prefix"`
	Token     string `json:"token"`
	LoginRate uint32 `json:"loginRate"`
}

type OB11Config struct {
	Network struct {
		HttpServers      []string    `json:"httpServers"`
		HttpClients      []string    `json:"httpClients"`
		WebsocketServers []string    `json:"websocketServers"`
		WebsocketClients []*WSClient `json:"websocketClients"`
	} `json:"network"`
	MusicSignUrl        string `json:"musicSignUrl"`
	EnableLocalFile2Url bool   `json:"enableLocalFile2Url"`
	ParseMultMsg        bool   `json:"parseMultMsg"`
}

type WSClient struct {
	Name              string `json:"name"`
	Enable            bool   `json:"enable"`
	URL               string `json:"url"`
	MessagePostFormat string `json:"messagePostFormat"`
	ReportSelfMessage bool   `json:"reportSelfMessage"`
	ReconnectInterval uint32 `json:"reconnectInterval"`
	Token             string `json:"token"`
	Debug             bool   `json:"debug"`
	HeartInterval     uint32 `json:"heartInterval"`
}

func NewOB11Config(mxid string, port uint32) *OB11Config {
	return &OB11Config{
		Network: struct {
			HttpServers      []string    `json:"httpServers"`
			HttpClients      []string    `json:"httpClients"`
			WebsocketServers []string    `json:"websocketServers"`
			WebsocketClients []*WSClient `json:"websocketClients"`
		}{
			HttpServers:      []string{},
			HttpClients:      []string{},
			WebsocketServers: []string{},
			WebsocketClients: []*WSClient{
				{
					Name:              mxid,
					Enable:            true,
					URL:               fmt.Sprintf("ws://localhost:%d", port),
					MessagePostFormat: "array",
					ReportSelfMessage: false,
					ReconnectInterval: 5000,
					Token:             mxid,
					Debug:             false,
					HeartInterval:     30000,
				},
			},
		},
		EnableLocalFile2Url: true,
		ParseMultMsg:        true,
	}
}
