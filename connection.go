package proxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type ConnectionPool map[string]*httputil.ClientConn

func (cp *ConnectionPool) Get(addr string) (*httputil.ClientConn, error) {
	if cc, ok := (*cp)[addr]; ok {
		return cc, nil
	}
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	cc := httputil.NewClientConn(c, nil)
	(*cp)[addr] = cc
	return cc, nil
}

func (cp *ConnectionPool) Release() {
	for _, cc := range *cp {
		cc.Close()
	}
}

func (p *Proxy) frontend_loop(c net.Conn, one_req chan *OnceRequest) {
	var proxy_srv *httputil.ClientConn
	var err error
	var req *http.Request
	var srv *Server

	cpool := make(ConnectionPool)

	proxy_cli := httputil.NewServerConn(c, nil)
	for {
		req, err = proxy_cli.Read()
		if err != nil {
			if err == httputil.ErrPersistEOF && req != nil {
				log.Println("[frontend_loop] ", "recv HTTP/1.0 request")
			} else {
				goto out
			}
		}
		fmt.Println("req: ", req)
		srv, err = p.getServerConn()
		if err != nil {
			log.Println("[frontend_loop] getServerConn error: ", err)
			goto out
		}

		proxy_srv, err = cpool.Get(srv.addr)
		if err != nil {
			goto out
		}

		one_req <- &OnceRequest{cp: &ConnPair{c: proxy_cli, s: proxy_srv}, req: req}
		err = proxy_srv.Write(req)
		if err != nil {
			goto out
		}
	}
out:
	log.Println("[frontend_loop] ", "out frontend_loop of error: ", err)
	one_req <- nil
	cpool.Release()
	c.Close()
}

func (p *Proxy) backend_loop(one_req chan *OnceRequest) {
	for {
		oq := <-one_req
		if oq == nil {
			break
		}
		resp, err := oq.cp.s.Read(oq.req)
		if err != nil {
			if err == httputil.ErrPersistEOF {
				log.Println("[backend_loop] ", err)
			} else {
				log.Println("[backend_loop] ", err)
			}
		}
		fmt.Println("resp: ", resp)
		err = oq.cp.c.Write(oq.req, resp)
		if err != nil {
			log.Println(err)
			return
		}
	}
	log.Println("[backend_loop] out backend_loop")
}

func (p *Proxy) getServerConn() (*Server, error) {
	item, err := p.Pool.Get()
	if err != nil {
		return nil, err
	}
	return item.(*Server), nil
}
