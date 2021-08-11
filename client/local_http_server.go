package main

import (
	. "SocketProxy/logger"
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type LocalHttpServer struct {
	listenAddr  string
	serverAddrs []string
	encType     uint8
	password    string
	timeout     time.Duration
}

func NewLocalHttpServer(localAddr string, servers []string, encType uint8, passwd string, timeout time.Duration) *LocalHttpServer {
	return &LocalHttpServer{
		listenAddr:  localAddr,
		serverAddrs: servers,
		encType:     encType,
		password:    passwd,
		timeout:     timeout,
	}
}
func (c *LocalHttpServer) getServerConn(remoteAddr string) (net.Conn, error) {
	for i := 0; i < len(c.serverAddrs); i++ {
		conn, err := DialServer(c.serverAddrs[i], remoteAddr, c.password, c.encType)
		if err == nil {
			return conn, nil
		}
	}
	return nil, errors.New("server connection failed")
}

func (s *LocalHttpServer) Run() {
	Logger.Info("start listen local http server ", s.listenAddr)
	err := http.ListenAndServe(s.listenAddr, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			newReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
			if err != nil {
				Logger.Error(err)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			newReq.Header = req.Header

			if v, ok := newReq.Header["Proxy-Connection"]; ok {
				newReq.Header["Connection"] = v
				delete(newReq.Header, "Proxy-Connection")
			}

			host, port, err := net.SplitHostPort(req.Host)
			if err != nil {
				host = req.Host
				port = "80"
			}

			// 域名解析
			ip, err := ParseDomain(host)
			if err != nil {
				Logger.Error(err)
				w.WriteHeader(http.StatusBadGateway)
				return
			}

			var conn net.Conn
			if !IPisProxy(ip) {
				conn, err = net.Dial("tcp", ip + ":" + port)
			} else {
				conn, err = s.getServerConn(ip + ":" + port)
			}
			if err != nil {
				w.WriteHeader(http.StatusRequestTimeout)
				return
			}
			defer conn.Close()

			err = newReq.Write(conn)
			if err != nil {
				Logger.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
			if err != nil {
				Logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			for k, vs := range resp.Header {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
			io.Copy(w, resp.Body)

		} else if req.Method == "CONNECT" {

			host, port, err := net.SplitHostPort(req.Host)
			if err != nil {
				host = req.Host
				port = "80"
			}
			// 域名解析
			ip, err := ParseDomain(host)
			if err != nil {
				Logger.Error(err)
				w.WriteHeader(http.StatusBadGateway)
				return
			}

			var conn net.Conn
			if !IPisProxy(ip) {
				conn, err = net.Dial("tcp", ip + ":" + port)
			} else {
				conn, err = s.getServerConn(ip + ":" + port)
			}
			if err != nil {
				w.WriteHeader(http.StatusRequestTimeout)
				return
			}
			defer conn.Close()

			w.WriteHeader(http.StatusOK)
			srcconn, bufrw, err := w.(http.Hijacker).Hijack()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer srcconn.Close()

			wg := &sync.WaitGroup{}
			wg.Add(2)
			//并发执行单元1: 将TCP连接拷贝到HTTP连接中
			go func() {
				defer wg.Done()
				//缓存处理
				n := bufrw.Reader.Buffered()
				if n > 0 {
					n64, err := io.CopyN(conn, bufrw, int64(n))
					if n64 != int64(n) || err != nil {
						Logger.Errorf("io.CopyN: %d %v\n", n64, err)
						return
					}
				}
				//进行全双工的双向数据拷贝（中继）
				io.Copy(conn, srcconn) //relay: src->dst
			}()
			//并发执行单元2：将HTTP连接拷贝到TCP连接中
			go func() {
				defer wg.Done()
				//进行全双工的双向数据拷贝（中继）
				io.Copy(srcconn, conn) //relay:dst->src
			}()
			wg.Wait()
		} else {
			w.WriteHeader(http.StatusBadGateway)
		}
	}))

	if err != nil {
		Logger.Fatal(err)
	}
}
