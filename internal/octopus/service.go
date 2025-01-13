package octopus

import (
	"fmt"
	"net/http"

	"github.com/duo/matrix-octopus-napcat/internal/common"
	"github.com/duo/matrix-octopus-napcat/internal/napcat"

	"github.com/duo/wsc"
	"github.com/rs/zerolog"
)

type Service struct {
	log zerolog.Logger

	bridge *wsc.Client

	manager *napcat.Manager
}

func NewService(log zerolog.Logger, config *common.Configure) *Service {
	options, err := wsc.NewClientOptions(
		config.Octopus.Addr,
		wsc.HTTPHeaders(http.Header{
			"Authorization":   []string{fmt.Sprintf("Basic %s", config.Octopus.Secret)},
			"X-Agent-ID":      []string{config.Octopus.ID},
			"X-Agent-Version": []string{"1"},
		}),
		wsc.PingTimeout(config.Octopus.PingInterval),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create octopus connection options")
	}

	service := &Service{
		log: log.With().Str("Service", "Octopus").Logger(),

		bridge: wsc.NewClient(options),
	}

	options.OnConnected = service.consumeWebsocket
	service.manager = napcat.NewManager(log, config, service.transmitFunc)

	return service
}

func (s *Service) Start() {
	s.log.Info().Msg("Starting Octopus service")

	if err := s.bridge.Connect(); err != nil {
		s.log.Fatal().Err(err).Msg("Failed to connect to octopus bridge")
	}
	s.manager.Start()
}

func (s *Service) Stop() {
	s.log.Info().Msg("Stopping Octopus service")

	s.manager.Stop()
	s.bridge.Disconnect()
}

func (s *Service) consumeWebsocket(_ *wsc.Client) {
	s.log.Info().Msg("Connected to Matrix-Octopus")

	for {
		var pkt common.Packet
		err := s.bridge.ReadJSON(&pkt)
		if err != nil {
			s.log.Warn().Err(err).Msg("Error reading from websocket")
			return
		}

		switch pkt.Type {
		case common.PktRequest:
			request := pkt.Payload.(*common.Request)
			s.log.Debug().
				Int64("packet", pkt.ID).
				Str("mxid", pkt.MXID).
				Msgf("Received packet %+v", request)
			go s.manager.ProcessRequest(pkt.ID, pkt.MXID, request)
		default:
			s.log.Warn().
				Int64("packet", pkt.ID).
				Str("mxid", pkt.MXID).
				Msgf("Packet type %s is not supported", pkt.Type)
		}
	}
}

func (s *Service) transmitFunc(pkt *common.Packet) error {
	err := s.bridge.WriteJSON(pkt)
	if err != nil {
		s.log.Warn().Err(err).
			Int64("packet", pkt.ID).
			Str("mxid", pkt.MXID).
			Msg("Failed to transmit packet")
	}

	return err
}
