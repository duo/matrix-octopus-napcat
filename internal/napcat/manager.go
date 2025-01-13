package napcat

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/duo/matrix-octopus-napcat/internal/common"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var (
	errMissingToken = common.ErrorResponse{
		HTTPStatus: http.StatusForbidden,
		Code:       "M_MISSING_TOKEN",
		Message:    "Missing authorization header",
	}
	errUnknownToken = common.ErrorResponse{
		HTTPStatus: http.StatusForbidden,
		Code:       "M_UNKNOWN_TOKEN",
		Message:    "Unknown authorization token",
	}

	upgrader = websocket.Upgrader{}
)

type Manager struct {
	log zerolog.Logger

	port           uint32
	path           string
	cmd            string
	initTimeout    time.Duration
	requestTimeout time.Duration

	clientIncPort uint32
	clients       map[string]*Client
	clientsLock   sync.Mutex

	transmitFunc func(*common.Packet) error

	server *http.Server
}

func NewManager(log zerolog.Logger, config *common.Configure, transmitFunc func(*common.Packet) error) *Manager {
	manager := &Manager{
		log: log.With().Str("Manager", "NapCat").Logger(),

		port:           config.NapCat.ManagerPort,
		path:           config.NapCat.Path,
		cmd:            config.NapCat.CMD,
		initTimeout:    config.NapCat.InitTimeout,
		requestTimeout: config.NapCat.RequestTimeout,

		clientIncPort: config.NapCat.ManagerPort + 1,
		clients:       make(map[string]*Client),

		transmitFunc: transmitFunc,
	}

	manager.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", manager.port),
		Handler: manager,
	}

	return manager
}

func (m *Manager) Start() {
	m.log.Info().Msgf("Starting NapCat manager, listening port %d", m.port)

	if err := m.cleanupConfig(); err != nil {
		m.log.Fatal().Err(err).Msg("Failed to cleanup NapCat configuration")
	}

	if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		m.log.Fatal().Err(err).Msg("Failed to start NapCat manager")
	}
}

func (m *Manager) Stop() {
	m.log.Info().Msg("Stopping NapCat manager")

	m.clientsLock.Lock()
	defer m.clientsLock.Unlock()

	for _, client := range m.clients {
		client.Disconnect()
	}
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		errMissingToken.Write(w)
		return
	}

	mxid := authHeader[len("Bearer "):]

	m.clientsLock.Lock()
	client, ok := m.clients[mxid]
	m.clientsLock.Unlock()
	if !ok {
		errUnknownToken.Write(w)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.log.Warn().Err(err).Msg("Failed to upgrade websocket request")
		return
	}

	client.updateConn(conn)
	client.Serve(m.transmitFunc)
}

func (m *Manager) ProcessRequest(id int64, mxid string, req *common.Request) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			m.log.Error().
				Int64("packet", id).
				Msgf("Panic while responding to command %s: %v\n%s", req.Type, panicErr, debug.Stack())
		}
	}()

	resp := m.handleRequest(mxid, req)

	respPkt := &common.Packet{
		ID:      id,
		MXID:    mxid,
		Type:    common.PktResponse,
		Payload: resp,
	}
	m.log.Debug().
		Int64("packet", id).
		Str("mxid", mxid).
		Msgf("Sending response %+v", respPkt)

	m.transmitFunc(respPkt)
}

func (m *Manager) handleRequest(mxid string, req *common.Request) *common.Response {
	client := m.getClient(mxid)

	switch req.Type {
	case common.MethodConnect:
		{
			m.clientsLock.Lock()
			defer m.clientsLock.Unlock()

			return genResponse(req.Type, nil, client.Connect())
		}
	case common.MethodDisconnect:
		{
			m.clientsLock.Lock()
			defer m.clientsLock.Unlock()

			client.Disconnect()
			delete(m.clients, mxid)

			return genResponse(req.Type, nil, nil)
		}
	case common.MethodLogin:
		ret, err := client.Login(req.Params.(*common.LoginStep))
		return genResponse(req.Type, ret, err)
	case common.MethodLogout:
		return genResponse(req.Type, nil, nil)
	case common.MethodGetLoginInfo:
		ret, err := client.GetLoginInfo()
		return genResponse(req.Type, ret, err)
	case common.MethodGetUserInfo:
		ret, err := client.GetUserInfo(req.Params.([]string)[0])
		return genResponse(req.Type, ret, err)
	case common.MethodGetGroupInfo:
		ret, err := client.GetGroupInfo(req.Params.([]string)[0])
		return genResponse(req.Type, ret, err)
	case common.MethodGetFriendList:
		ret, err := client.GetFriendList()
		return genResponse(req.Type, ret, err)
	case common.MethodGetGroupList:
		ret, err := client.GetFriendList()
		return genResponse(req.Type, ret, err)
	case common.MethodGetGroupMemberList:
		ret, err := client.GetGroupMemberList(req.Params.([]string)[0])
		return genResponse(req.Type, ret, err)
	case common.MethodGetGroupMemberInfo:
		ret, err := client.GetGroupMemberInfo(req.Params.([]string)[0], req.Params.([]string)[1])
		return genResponse(req.Type, ret, err)
	case common.MethodSendMessage:
		ret, err := client.SendMessage(req.Params.(*common.Message))
		return genResponse(req.Type, ret, err)
	case common.MethodRevokeMessage:
		return genResponse(req.Type, nil, client.RevokeMessage(req.Params.([]string)[0]))
	default:
		return genResponse(req.Type, nil, fmt.Errorf("Method %s is not supported", req.Type))
	}
}

func (m *Manager) getClient(mxid string) *Client {
	m.clientsLock.Lock()
	defer m.clientsLock.Unlock()

	client, ok := m.clients[mxid]
	if ok {
		return client
	} else {
		client = NewClient(m.log, mxid, m.path, m.port, m.clientIncPort, m.cmd, m.initTimeout, m.requestTimeout)
		m.clients[mxid] = client
		m.clientIncPort++
		return client
	}
}

func genResponse(mType common.MethodType, data any, err error) *common.Response {
	resp := &common.Response{
		Type: mType,
		Data: data,
	}

	if err != nil {
		resp.Error = &common.ErrorResponse{
			Code:    "PROCESS_FAILED",
			Message: err.Error(),
		}
	}

	return resp
}

func (m *Manager) cleanupConfig() error {
	return filepath.Walk(m.path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			filename := filepath.Base(path)
			if strings.HasPrefix(filename, "napcat_") || strings.HasPrefix(filename, "onebot11_") {
				if err := os.Remove(path); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
