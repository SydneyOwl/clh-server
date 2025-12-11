package server

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/internal/msg"
	"github.com/sydneyowl/clh-server/internal/verifier"
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/config"
	msg2 "github.com/sydneyowl/clh-server/pkg/msg"
	"github.com/sydneyowl/clh-server/pkg/trans"
)

const (
	connReadTimeout time.Duration = 10 * time.Second
)

type Service struct {
	cfg        *config.Config
	tlsEnabled bool
	tlsConfig  *tls.Config
	ctx        context.Context
	cancel     context.CancelFunc
	listener   net.Listener
	verifier   verifier.Verifier
}

func NewService(cfg *config.Config) (*Service, error) {
	svr := &Service{
		cfg: cfg,
	}

	svr.tlsEnabled = cfg.Message.EnableTLS
	svr.verifier = verifier.NewAuthKeyVerifier(svr.cfg.Message.Key)

	if svr.tlsEnabled {
		tlscfg, err := trans.NewServerTLSConfig(cfg.Message.TLSCertPath, cfg.Message.TLSKeyPath, cfg.Message.TLSCACertPath)
		if err != nil {
			return nil, err
		}
		svr.tlsConfig = tlscfg
	}

	return svr, nil
}

func (svr *Service) Run(ctx context.Context) {
	var (
		ln  net.Listener
		err error
	)
	ctx, cancel := context.WithCancel(ctx)
	svr.ctx = ctx
	svr.cancel = cancel

	address := net.JoinHostPort(svr.cfg.Server.BindAddr, strconv.Itoa(svr.cfg.Server.BindPort))
	if svr.tlsEnabled {
		ln, err = tls.Listen("tcp", address, svr.tlsConfig)
	} else {
		ln, err = net.Listen("tcp", address)
	}
	if err != nil {
		slog.Panicf("failed to listen on %s: %w", address, err)
	}
	svr.listener = ln

	slog.Infof("Server listening on %s", address)

	svr.handleListener(svr.listener)
	<-svr.ctx.Done()
	if svr.listener != nil {
		svr.Shutdown()
	}
}

func (svr *Service) Shutdown() {
	if svr.listener != nil {
		_ = svr.listener.Close()
	}
	if svr.cancel != nil {
		svr.cancel()
	}
}

func (svr *Service) handleListener(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			slog.Warnf("listener for incoming connections from client closed")
			return
		}

		// dispatcher works here
		go svr.handleConn(c)
	}
}

func (svr *Service) handleConn(conn net.Conn) {
	var (
		rawMsg msg2.Message
		err    error
	)

	_ = conn.SetReadDeadline(time.Now().Add(connReadTimeout))

	if rawMsg, err = msg.ReadMsg(conn); err != nil {
		slog.Debugf("failed to read message: %v", err)
		_ = conn.Close()
		return
	}

	// read infinitely.
	_ = conn.SetReadDeadline(time.Time{})

	switch m := rawMsg.(type) {
	case *msgproto.HandshakeRequest:
		slog.Debugf("received handshake request from client: %s %s", m.ClientType, m.Ver)
		if err := svr.verifier.VerifyLogin(m); err != nil {
			slog.Warnf("failed to verify login: %v", err)
			_ = msg.WriteMsg(conn, &msgproto.HandshakeResponse{
				RunId:  m.RunId,
				Accept: false,
				Error:  err.Error(),
			})
			_ = conn.Close()
			return
		}
		// accept connection
		slog.Infof("client login: ip [%s] os [%s] type [%s]", conn.RemoteAddr().String(), m.Os, m.ClientType)
		_ = msg.WriteMsg(conn, &msgproto.HandshakeResponse{
			RunId:  m.RunId,
			Accept: true,
			Error:  "",
		})
		ctl := NewControl(svr.ctx, m.RunId, conn)
		ctl.Run()
	default:
		slog.Debugf("received a request that cannot be understood by pre-handler.")
	}
}
