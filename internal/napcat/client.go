package napcat

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/duo/matrix-octopus-napcat/internal/common"

	"github.com/go-cmd/cmd"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	lru "github.com/hashicorp/golang-lru/v2"
)

const (
	remoteURL = "http://localhost"

	LoginStepQR       = "me.lxduo.octopus.login.qr"
	LoginStepWait     = "me.lxduo.octopus.login.wait"
	LoginStepComplete = "me.lxduo.octopus.login.complete"

	defaultAvatar = "bad9cbb852b22fe58e62f3f23c7d63d2"
)

var (
	avatarSizes = []int{0, 640, 140, 100, 41, 40}
	lruCache    *lru.Cache[int64, string]
	once        sync.Once
)

type Client struct {
	log zerolog.Logger

	mxid        string
	path        string
	managerPort uint32
	port        uint32
	uin         int64

	instance     *cmd.Cmd
	instanceLock sync.Mutex

	credential string
	client     *resty.Client

	command        string
	initTimeout    time.Duration
	requestTimeout time.Duration

	conn      *websocket.Conn
	writeLock sync.Mutex

	websocketRequests     map[string]chan<- *Response
	websocketRequestsLock sync.RWMutex
	websocketRequestID    int64

	mutex common.KeyMutex
}

func NewClient(log zerolog.Logger, mxid, path string, managerPort, port uint32, command string, initTimeout, requestTimeout time.Duration) *Client {
	return &Client{
		log: log.With().Str("Client", mxid).Uint32("Port", port).Logger(),

		mxid:           mxid,
		path:           path,
		managerPort:    managerPort,
		port:           port,
		command:        command,
		initTimeout:    initTimeout,
		requestTimeout: requestTimeout,

		client:            resty.New(),
		websocketRequests: make(map[string]chan<- *Response),

		mutex: common.NewHashed(47),
	}
}

func (c *Client) Serve(transmitFunc func(*common.Packet) error) {
	defer func() {
		c.log.Info().Msg("Client disconnected")
		c.updateConn(nil)
	}()

	for {
		var m map[string]interface{}
		if err := c.conn.ReadJSON(&m); err != nil {
			c.log.Warn().Err(err).Msg("Error reading from connection")
			break
		}

		c.log.Debug().Msgf("Receive NapCat payload: %+v", m)

		payload, err := UnmarshalPayload(m)
		if err != nil {
			c.log.Warn().Err(err).Msg("Failed to unmarshal payload")
			continue
		}

		switch payload.PayloadType() {
		case PaylaodRequest:
			c.log.Warn().Msgf("Request %s not supported", payload.(*Request).Action)
		case PayloadResponse:
			go c.handleResponse(payload.(*Response))
		case PayloadEvent:
			go c.handleEvent(payload.(IEvent), transmitFunc)
		}
	}
}

func (c *Client) genCompleteStep() (*common.LoginStep, error) {
	if resp, err := c.getLoginInfo(); err != nil {
		return nil, err
	} else {
		return &common.LoginStep{
			StepID:       LoginStepComplete,
			Type:         common.LoginStepTypeComplete,
			Instructions: fmt.Sprintf("Successfully logged in as %s", resp.Data.Nick),
			CompleteParams: &common.LoginCompleteParams{
				LoginInfo: &common.UserInfo{
					ID:   resp.Data.Uin,
					Name: resp.Data.Nick,
				},
			},
		}, nil
	}
}

func (c *Client) updateConn(conn *websocket.Conn) {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}
	c.conn = conn
}

func (c *Client) overwriteWebUIConfig() error {
	config := &WebUIConfig{
		Host:      "localhost",
		Port:      c.port,
		Prefix:    "",
		Token:     c.mxid,
		LoginRate: 7,
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	configPath := filepath.Join(c.path, "webui.json")
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)

	return err
}

func (c *Client) checkLogin() (*CheckLoginResponse, error) {
	resp, err := c._checkLogin()

	if err != nil {
		if _, ok := err.(*UnauthorizedError); ok {
			c.auth()
			return c._checkLogin()
		}
	}

	return resp, err
}

func (c *Client) updateOB11Config() error {
	if err := c._updateOB11Config(); err != nil {
		if _, ok := err.(*UnauthorizedError); ok {
			c.auth()
			return c._updateOB11Config()
		}

		return err
	}

	return nil
}

func (c *Client) getLoginInfo() (*GetLoginInfoResponse, error) {
	resp, err := c._getLoginInfo()

	if err != nil {
		if _, ok := err.(*UnauthorizedError); ok {
			c.auth()
			return c._getLoginInfo()
		}
	}

	return resp, err
}

func (c *Client) auth() error {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(AuthRequest{Token: c.mxid}).
		SetResult(&AuthResponse{}).
		Post(fmt.Sprintf("%s:%d/api/auth/login", remoteURL, c.port))

	if err != nil {
		return err
	}

	response := resp.Result().(*AuthResponse)
	if response.Code != 0 {
		return errors.New(response.Message)
	}

	c.credential = response.Data.Credential

	return nil
}

func (c *Client) _checkLogin() (*CheckLoginResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.credential)).
		SetResult(&CheckLoginResponse{}).
		Post(fmt.Sprintf("%s:%d/api/QQLogin/CheckLoginStatus", remoteURL, c.port))

	if err != nil {
		return nil, err
	}

	response := resp.Result().(*CheckLoginResponse)
	if response.Code != 0 {
		if response.Message == "Unauthorized" {
			return nil, &UnauthorizedError{Message: response.Message}
		} else {
			return nil, errors.New(response.Message)
		}
	}

	return response, nil
}

func (c *Client) _updateOB11Config() error {
	config := NewOB11Config(c.mxid, c.managerPort)
	configJson, err := json.Marshal(config)
	if err != nil {
		return err
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.credential)).
		SetBody(&SetOB11ConfigRequest{Config: string(configJson)}).
		SetResult(&APIResponse{}).
		Post(fmt.Sprintf("%s:%d/api/OB11Config/SetConfig", remoteURL, c.port))

	if err != nil {
		return err
	}

	response := resp.Result().(*APIResponse)
	if response.Code != 0 {
		if response.Message == "Unauthorized" {
			return &UnauthorizedError{Message: response.Message}
		} else {
			return errors.New(response.Message)
		}
	}

	return nil
}

func (c *Client) _getLoginInfo() (*GetLoginInfoResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.credential)).
		SetResult(&GetLoginInfoResponse{}).
		Post(fmt.Sprintf("%s:%d/api/QQLogin/GetQQLoginInfo", remoteURL, c.port))

	if err != nil {
		return nil, err
	}

	response := resp.Result().(*GetLoginInfoResponse)
	if response.Code != 0 {
		if response.Message == "Unauthorized" {
			return nil, &UnauthorizedError{Message: response.Message}
		} else {
			return nil, errors.New(response.Message)
		}
	}

	return response, nil
}

func (c *Client) request(req *Request) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.requestTimeout)
	defer cancel()

	req.Echo = fmt.Sprint(atomic.AddInt64(&c.websocketRequestID, 1))

	respChan := make(chan *Response, 1)

	c.addWebsocketResponseWaiter(req.Echo, respChan)
	defer c.removeWebsocketResponseWaiter(req.Echo, respChan)

	c.log.Debug().Str("echo", req.Echo).Str("action", req.Action).Msgf("Send NapCat request %+v", req)
	if err := c._request(req); err != nil {
		return nil, err
	}

	select {
	case resp := <-respChan:
		if resp.Status != "ok" {
			return resp, fmt.Errorf("%s NapCat response retcode: %d", resp.Status, resp.Retcode)
		} else {
			return resp.Data, nil
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) _request(req *Request) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	if c.conn == nil {
		return errors.New("websocket not connected")
	}

	_ = c.conn.SetWriteDeadline(time.Now().Add(c.requestTimeout))
	return c.conn.WriteJSON(req)
}

func (c *Client) addWebsocketResponseWaiter(echo string, waiter chan<- *Response) {
	c.websocketRequestsLock.Lock()
	c.websocketRequests[echo] = waiter
	c.websocketRequestsLock.Unlock()
}

func (c *Client) removeWebsocketResponseWaiter(echo string, waiter chan<- *Response) {
	c.websocketRequestsLock.Lock()
	existingWaiter, ok := c.websocketRequests[echo]
	if ok && existingWaiter == waiter {
		delete(c.websocketRequests, echo)
	}
	c.websocketRequestsLock.Unlock()
	close(waiter)
}

func getUserAvatarURL(uin int64) string {
	if url, ok := getAvatarCache().Get(uin); ok {
		return url
	}

	url := ""
	for _, size := range avatarSizes {
		url = fmt.Sprintf("https://q.qlogo.cn/headimg_dl?dst_uin=%d&spec=%d", uin, size)
		data, err := GetBytes(url)
		if err != nil || fmt.Sprintf("%x", md5.Sum(data)) == defaultAvatar {
			continue
		} else {
			break
		}
	}
	getAvatarCache().Add(uin, url)

	return url
}

func getGroupAvatarURL(groupId int64) string {
	return fmt.Sprintf("https://p.qlogo.cn/gh/%d/%d/0", groupId, groupId)
}

func getAvatarCache() *lru.Cache[int64, string] {
	once.Do(func() {
		lruCache, _ = lru.New[int64, string](1024)
	})
	return lruCache
}
