package grpcquic

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	quic "github.com/lucas-clemente/quic-go"
)

///////////////////////////////////////////////////////////////////////////////
// Connection

var _ net.Conn = (*Conn)(nil)

type Conn struct {
	sess   quic.Session
	stream quic.Stream
}

func NewConn(sess quic.Session) (net.Conn, error) {
	stream, err := sess.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	return &Conn{sess, stream}, nil
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.stream.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.stream.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) Close() error {
	// @TODO: log this
	c.stream.Close()

	return c.sess.CloseWithError(0, "")
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.sess.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.sess.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *Conn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)

}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

///////////////////////////////////////////////////////////////////////////////
// Listener

var _ net.Listener = (*Listener)(nil)

type Listener struct {
	ql quic.Listener
}

func Listen(ql quic.Listener) net.Listener {
	return &Listener{ql}
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	sess, err := l.ql.Accept(context.Background())
	if err != nil {
		return nil, err
	}

	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}

	return &Conn{sess, stream}, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	return l.ql.Close()
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.ql.Addr()
}

///////////////////////////////////////////////////////////////////////////////
// Dialer

var quicConfig = &quic.Config{
	// MaxReceiveStreamFlowControlWindow:     3 * (1 << 20),   // 3 MB
	// MaxReceiveConnectionFlowControlWindow: 4.5 * (1 << 20), // 4.5 MB
	// Versions: []quic.VersionNumber{101},
	// AcceptCookie: func(clientAddr net.Addr, cookie *quic.Cookie) bool {
	// 	// TODO(#6): require source address validation when under load
	// 	return true
	// },
	KeepAlive: true,
}

func NewPacketConn(addr string) (net.PacketConn, error) {
	// create a packet conn for outgoing connections
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	return net.ListenUDP("udp", udpAddr)
}

func NewQuicDialer(tlsConf *tls.Config) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, target string) (net.Conn, error) {
		sess, err := quic.DialAddrContext(ctx, target, tlsConf, quicConfig)
		if err != nil {
			return nil, err
		}

		return NewConn(sess)
	}
}
