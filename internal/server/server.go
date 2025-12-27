package server

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"

	"github.com/sydneyowl/clh-server/internal/cache"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/internal/msg"
	"github.com/sydneyowl/clh-server/internal/verifier"
	"github.com/sydneyowl/clh-server/pkg/config"
	msg2 "github.com/sydneyowl/clh-server/pkg/msg"
	"github.com/sydneyowl/clh-server/pkg/trans"
)

const (
	connReadTimeout time.Duration = 10 * time.Second
)

type Service struct {
	cfg            *config.Config
	ctrlManager    *ControlManager
	clientRegistry *ClientRegistry
	cache          cache.CLHCache
	tlsEnabled     bool
	tlsConfig      *tls.Config
	ctx            context.Context
	cancel         context.CancelFunc
	listener       net.Listener
	verifier       verifier.Verifier
}

func NewService(cfg *config.Config) (*Service, error) {
	svr := &Service{
		cfg:            cfg,
		ctrlManager:    NewControlManager(),
		clientRegistry: NewClientRegistry(0),
		cache:          cache.NewMemoryCache(),
		tlsEnabled:     cfg.Server.Encrypt.EnableTLS,
		verifier:       verifier.NewAuthKeyVerifier(cfg.Server.Encrypt.Key),
	}

	if svr.tlsEnabled {
		tlscfg, err := trans.NewServerTLSConfig(cfg.Server.Encrypt.TLSCertPath, cfg.Server.Encrypt.TLSKeyPath, cfg.Server.Encrypt.TLSCACertPath)
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
		slog.Panicf("failed to listen on %s: %s", address, err)
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
	case *clh_proto.HandshakeRequest:
		slog.Debugf("received handshake request from client: %s %s", m.ClientType, m.Ver)
		if err := svr.verifier.VerifyLogin(m); err != nil {
			slog.Warnf("failed to verify login: %v", err)
			_ = msg.WriteMsg(conn, &clh_proto.HandshakeResponse{
				RunId:  m.RunId,
				Accept: false,
				Error:  err.Error(),
			})
			_ = conn.Close()
			return
		}

		clientType, err := ValidateClientType(m.ClientType)
		if err != nil {
			slog.Warnf("failed to validate client type: %v", err)
			_ = msg.WriteMsg(conn, &clh_proto.HandshakeResponse{
				RunId:  m.RunId,
				Accept: false,
				Error:  err.Error(),
			})
			_ = conn.Close()
			return
		}

		// accept connection
		slog.Infof("client login: ip [%s] os [%s] type [%s]", conn.RemoteAddr().String(), m.Os, m.ClientType)
		_ = msg.WriteMsg(conn, &clh_proto.HandshakeResponse{
			RunId:  m.RunId,
			Accept: true,
			Error:  "",
		})
		// create control with a lightweight callback to get current senders
		ctl := NewControl(svr.cfg, m.RunId, clientType, svr.cache, conn, svr.ctx, svr.clientRegistry)
		// pop out dupe conn
		if old := svr.ctrlManager.Add(m.RunId, ctl); old != nil {
			old.WaitForQuit()
			slog.Infof(" [%s] replaced by new client.", old.runID)
		}
		// register client in registry
		svr.clientRegistry.Register(ClientInfo{RunID: m.RunId, Type: clientType, Addr: conn.RemoteAddr().String()})

		go ctl.Run()
		go func() {
			ctl.WaitForQuit()
			svr.ctrlManager.Del(m.RunId, ctl)
			// unregister from registry
			svr.clientRegistry.Unregister(m.RunId)
			slog.Tracef("Released %s from ctrl manager.", m.RunId)
		}()
	default:
		slog.Debugf("received a request that cannot be understood by pre-handler.")
	}
}
