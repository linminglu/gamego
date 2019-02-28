package room

import (
	"io"
	"net"
	"os"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go"

	. "local.com/abc/game/msg"
)

var (
	connectCount int32        //当前连接数
	maxConnect   int32 = 4000 //最大连接数
)

func startServer(config *AppConfig) {
	if config.Tcp.Listen != "" {
		go tcpServer(config)
	}
	if config.Udp.Listen != "" {
		go udpServer(config)
	}
}

func tcpServer(config *AppConfig) {
	// resolve address & start listening
	l, err := net.Listen("tcp", config.Tcp.Listen)
	log.Info("listening on:", l.Addr())
	checkError(err)

	lis := l.(*net.TCPListener)
	defer lis.Close()
	// loop accepting
	for {
		select {
		case <-signal.Die():
			return
		default:
			conn, err := lis.AcceptTCP()
			if err != nil {
				log.Warning("accept failed:", err)
				continue
			}
			atomic.LoadInt32(&connectCount)
			// set socket read buffer
			conn.SetReadBuffer(config.Tcp.ReadBuf)
			// set socket write buffer
			conn.SetWriteBuffer(config.Tcp.WriteBuf)
			go handleClient(conn)
		}
	}
}

func udpServer(config *AppConfig) {
	l, err := kcp.Listen(config.Udp.Listen)
	checkError(err)
	log.Info("udp listening on:", l.Addr())

	lis := l.(*kcp.Listener)
	defer lis.Close()
	if err := lis.SetReadBuffer(config.Udp.ReadBuf); err != nil {
		log.Println("SetReadBuffer:", err)
	}
	if err := lis.SetWriteBuffer(config.Udp.WriteBuf); err != nil {
		log.Println("SetWriteBuffer:", err)
	}
	if err := lis.SetDSCP(config.Udp.Dscp); err != nil {
		log.Println("SetDSCP:", err)
	}
	// loop accepting
	for {
		select {
		case <-signal.Die():
			return
		default:
			conn, err := lis.AcceptKCP()
			if err != nil {
				log.Warning("accept failed:", err)
				continue
			}
			atomic.LoadInt32(&connectCount)
			// set kcp parameters
			conn.SetWindowSize(config.Kcp.Sndwnd, config.Kcp.Rcvwnd)
			conn.SetNoDelay(config.Kcp.Nodelay, config.Kcp.Interval, config.Kcp.Resend, config.Kcp.Nc)
			conn.SetStreamMode(true)
			conn.SetMtu(config.Kcp.Mtu)
			// start a goroutine for every incoming connection for reading
			go handleClient(conn)
		}
	}
}

// handleClient
func handleClient(conn net.Conn) {
	defer func() {
		atomic.AddInt32(&connectCount, -1)
		conn.Close()
	}()
	count := atomic.AddInt32(&connectCount, 1)
	if count <= maxConnect {
		newSession(conn)
	}
}

// create a new session object for the connection
func newSession(conn net.Conn) {
	signal.Add(1)
	defer signal.Add(-1)

	// 读取头信息
	head := [16]byte{}
	// 读取用户信息
	n, err := io.ReadFull(conn, head[:])
	if err != nil || n != 16 {
		return
	}

	agent, uid, ip := GetUserHead(head[:])
	sess := &Session{
		Id:       agent,
		UserId:   uid,
		Ip:       ip,
		Created:  time.Now(),
		dieChan:  make(chan struct{}),
		sendChan: make(chan interface{}, 512),
		stopSend: make(chan struct{}),
	}

	log.Infof("connection agent:%v, user:%v, Ip:%v", agent, uid, ip)
	// cleanup work
	defer func() {
		// 连接断开事件
		Call(func() { userOffline(sess) })
		sess.Close()
		log.Infof("connection closed user:%v, Ip:%v", uid, ip)
	}()

	// 连接事件
	Call(func() { userOnline(sess) })
	sess.Start(&NetStream{
		conn: conn,
	})
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}

type NetStream struct {
	conn net.Conn
	head [HeadLen]byte
}

func (stream *NetStream) Send(d []byte) error {
	if _, e := stream.conn.Write(d); e == nil {
		return nil
	} else {
		return e
	}
}

func (stream *NetStream) Recv() ([]byte, error) {
	head := stream.head[:]
	n, err := io.ReadFull(stream.conn, head)
	if err != nil || n != HeadLen {
		return nil, err
	}
	size := int(GetHeadLen(head))
	payload := make([]byte, HeadLen+size)
	n, err = io.ReadFull(stream.conn, payload[HeadLen:])
	if err != nil || n != size {
		return nil, err
	}
	copy(payload[:HeadLen], head)
	return payload, nil
}

func (stream *NetStream) Close() {
	stream.conn.Close()
}