package main

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net"
	"net/http"
	"os"
)

const socks5Ver = 0x05
const cmdConnect = 0x01
const cmdUdp = 0x03
const atypIPV4 = 0x01
const atypeHOST = 0x03
const atypeIPV6 = 0x04
const pathtcp = "/tcp"
const pathudp = "/udp"
const BUFLEN = 65536
const CONNS = 65536
const opcode = 0xFF

var conntable [CONNS]net.Conn

func main() {
	addr := ":" + os.Getenv("PORT")
	log.Println("listen to Port: ", addr)
	http.Handle(pathtcp, websocket.Handler(ProcessTcp))
	http.Handle(pathudp, websocket.Handler(ProcessUdp))
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func ProcessUdp(conn *websocket.Conn) {
	var msg [BUFLEN]byte
	var addr string
	defer conn.Close()
	for {
		n, err := conn.Read(msg[0:])
		if err != nil {
			log.Println("Read Msg:", err)
			break
		}
		if msg[0] == 0x01 {
			log.Println("Keep UDP Alive")
			continue
		}
		//xor with opcode
		for j := 9; j < n; j++ {
			msg[j] = msg[j] ^ opcode
		}
		atyp := msg[9+3]
		switch atyp {
		case atypIPV4:
			addr = fmt.Sprintf("%d.%d.%d.%d", msg[9+4], msg[9+5], msg[9+6], msg[9+7])
			log.Println("UDP: atyp is IPV4:", addr)
		case atypeHOST:
			log.Println("UDP: atyp is HOST")
			continue
		case atypeIPV6:
			log.Println("UDP: atyp is IPV6")
			continue
		default:
			continue
		}
		port := binary.BigEndian.Uint16(msg[9+8 : 9+10])
		udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", addr, port))
		if err != nil {
			log.Println("ResolveUDPAddr", err)
			continue
		}
		go ProcessUdpreply(udpAddr, msg, n, conn)
	}
}

func ProcessUdpreply(udpAddr *net.UDPAddr, msg [BUFLEN]byte, n int, conn *websocket.Conn) {
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Println("DialUDP", err)
		return
	}
	udpConn.Write(msg[9+10 : n])
	n, err = udpConn.Read(msg[9+10:])
	if err != nil {
		log.Println("ReadUDP", err)
		return
	}
	udpConn.Close()
	//xor with opcode
	for j := 9; j < 9+10+n; j++ {
		msg[j] = msg[j] ^ opcode
	}
	//send to client
	_, err = conn.Write(msg[:9+10+n])
	if err != nil {
		log.Println("WriteMsg", err)
	}
}

func ProcessTcp(conn *websocket.Conn) {
	defer conn.Close()
	var buf [BUFLEN]byte
	for {
		//read websocket msg
		n, err := conn.Read(buf[0:])
		if err != nil {
			log.Println("ProcessTcp.Read", err)
			return
		}
		//process flag in header
		if buf[0] == 0x01 {
			log.Println("Keep TCP Alive")
			continue
		}
		//get entire header
		for uint32(n) < 8 {
			n1, err := conn.Read(buf[n:8])
			if err != nil {
				log.Println("ProcessTcp.Read:", err)
				return
			}
			n = n + n1
		}
		//get data length and entire data
		length := binary.BigEndian.Uint32(buf[4:])
		for uint32(n) < length {
			n1, err := conn.Read(buf[n:length])
			if err != nil {
				log.Println("ProcessTcp.Read:", err)
				return
			}
			n = n + n1
		}
		//xor with opcode
		for j := 8; j < n; j++ {
			buf[j] = buf[j] ^ opcode
		}
		//process type in header
		switch buf[3] {
		case 0x01:
			go auth1(buf, conn)
		case 0x02:
			go auth2(buf, conn)
		case 0x00:
			sendtoserver(buf, n)
		}
	}
}

func auth1(buf [BUFLEN]byte, conn *websocket.Conn) {
	// +----+----------+----------+
	// |VER | NMETHODS | METHODS  |
	// +----+----------+----------+
	// | 1  |    1     | 1 to 255 |
	// +----+----------+----------+
	// VER: 协议版本，socks5为0x05
	// NMETHODS: 支持认证的方法数量
	// METHODS: 对应NMETHODS，NMETHODS的值为多少，METHODS就有多少个字节。RFC预定义了一些值的含义，内容如下:
	// X’00’ NO AUTHENTICATION REQUIRED
	// X’02’ USERNAME/PASSWORD

	if buf[8] != socks5Ver {
		log.Printf("auth1: not supported ver:%d\n", buf[8])
		return
	}

	// +----+--------+
	// |VER | METHOD |
	// +----+--------+
	// | 1  |   1    |
	// +----+--------+
	buf[9] = 0x00
	//xor with opcade
	buf[8] = buf[8] ^ opcode
	buf[9] = buf[9] ^ opcode
	//set data length
	binary.BigEndian.PutUint32(buf[4:], 10)
	_, err := conn.Write(buf[0:10])
	if err != nil {
		log.Println("auth1.Write:", err)
	}
}

func auth2(buf [BUFLEN]byte, conn *websocket.Conn) {
	// +----+-----+-------+------+----------+----------+
	// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	// +----+-----+-------+------+----------+----------+
	// | 1  |  1  | X'00' |  1   | Variable |    2     |
	// +----+-----+-------+------+----------+----------+
	// VER 版本号，socks5的值为0x05
	// CMD 0x01表示CONNECT请求
	// RSV 保留字段，值为0x00
	// ATYP 目标地址类型，DST.ADDR的数据对应这个字段的类型。
	//   0x01表示IPv4地址，DST.ADDR为4个字节
	//   0x03表示域名，DST.ADDR是一个可变长度的域名
	// DST.ADDR 一个可变长度的值
	// DST.PORT 目标端口，固定2个字节

	ver, cmd, atyp := buf[8], buf[9], buf[11]
	if ver != socks5Ver {
		log.Printf("auth2: not supported ver:%d\n", ver)
		return
	}
	if cmd != cmdConnect && cmd != cmdUdp {
		log.Printf("auth2: not supported cmd:%d\n", cmd)
		return
	}
	var addr string
	var port uint16
	switch atyp {
	case atypIPV4:
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[12], buf[13], buf[14], buf[15])
		port = binary.BigEndian.Uint16(buf[16 : 16+2])
	case atypeHOST:
		hostSize := uint(buf[12])
		addr = string(buf[13 : 13+hostSize])
		port = binary.BigEndian.Uint16(buf[13+hostSize : 13+hostSize+2])
	case atypeIPV6:
		log.Println("IPv6: no supported yet")
		return
	default:
		log.Println("invalid atyp")
		return
	}

	// +----+-----+-------+------+----------+----------+
	// |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	// +----+-----+-------+------+----------+----------+
	// | 1  |  1  | X'00' |  1   | Variable |    2     |
	// +----+-----+-------+------+----------+----------+
	// VER socks版本，这里为0x05
	// REP Relay field,内容取值如下 X’00’ succeeded
	// RSV 保留字段
	// ATYPE 地址类型
	// BND.ADDR 服务绑定的地址
	// BND.PORT 服务绑定的端口DST.PORT
	buf[8], buf[9], buf[10], buf[11], buf[12], buf[13], buf[14], buf[15], buf[16], buf[17] = 0x05, 0x00, 0x00, 0x01, 127, 0, 0, 1, 4, 57
	//xor with opcode
	for j := 8; j < 18; j++ {
		buf[j] = buf[j] ^ opcode
	}
	//set data length
	binary.BigEndian.PutUint32(buf[4:], 18)
	if cmd == cmdConnect {
		connIndex := binary.BigEndian.Uint16(buf[1:])
		dest, err := net.Dial("tcp", fmt.Sprintf("%v:%v", addr, port))
		if err != nil {
			log.Println("auth2.Dial:", err)
			return
		}
		log.Println("dial", addr, port)
		conntable[connIndex] = dest
		go readfromserver(conn, dest, connIndex)
	}
	_, err := conn.Write(buf[0:18])
	if err != nil {
                log.Println("auth2.Write:", err)
		return
	}
}

func readfromserver(ws *websocket.Conn, s net.Conn, connIndex uint16) {
	var buf [BUFLEN]byte
	buf[0] = 0x00
	binary.BigEndian.PutUint16(buf[1:], connIndex)
	log.Println("readfromserver.connIndex:", connIndex)
	buf[3] = 0x00
	for {
		//read from server
		n, err := s.Read(buf[8:])
		if err != nil {
			log.Println("readfromserver.Read:", err)
			return
		}
		//set data length
		binary.BigEndian.PutUint32(buf[4:], uint32(8+n))
		log.Println("readfromserver length:", n)
		//xor with opcode
		for j := 8; j < 8+n; j++ {
			buf[j] = buf[j] ^ opcode
		}
		//send to websocket
		_, err = ws.Write(buf[0 : 8+n])
		if err != nil {
			log.Println("readfromserver.Write:", err)
			return
		}
	}
}

func sendtoserver(buf [BUFLEN]byte, n int) {
	connIndex := binary.BigEndian.Uint16(buf[1:])
	log.Println("sendtoserver.connIndex:", connIndex)
	conn := conntable[connIndex]
	_, err := conn.Write(buf[8:n])
	if err != nil {
		log.Println("sendtoserver.Write:", err)
	}
}
