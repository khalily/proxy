package proxy

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)
import "errors"

import "fmt"

import "net"
import "os"

type JsonConfig map[string]interface{}

func NewJsonConfig() *JsonConfig {
	jsonConfig := make(JsonConfig)
	return &jsonConfig
}

func (cfg *JsonConfig) ReadCfgFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	buffer := make([]byte, fi.Size())
	n, err := file.Read(buffer)
	if err != nil || n != len(buffer) {
		return err
	}
	err = json.Unmarshal(buffer, cfg)
	return err
}

func (cfg *JsonConfig) GetStringOf(name string) (string, error) {
	if v, ok := (*cfg)[name]; ok {
		switch value := v.(type) {
		case string:
			return string(value), nil
		}
	}
	return "", errors.New(fmt.Sprintf("error: not found string of name %s", name))
}

func (cfg *JsonConfig) GetStringsOf(name string) ([]string, error) {
	if v, ok := (*cfg)[name]; ok {
		switch value := v.(type) {
		case []interface{}:
			var strs []string
			for _, s := range value {
				strs = append(strs, s.(string))
			}
			return strs, nil
		default:
			fmt.Printf("value of name %s is %v and type %T\n", name, value, value)
		}
	}
	return nil, errors.New(fmt.Sprintf("error: not found strings of name %s", name))
}

type Server struct {
	addr string
}

type ServerPool struct {
	pool map[string]*Server
}

func (sl *ServerPool) Get(addr string) (*Server, error) {
	if srv, ok := sl.pool[addr]; ok {
		return srv, nil
	}
	return nil, errors.New(fmt.Sprintf("not found server addr %s", addr))
}

func (sl *ServerPool) Put(srv *Server) {
	sl.pool[srv.addr] = srv
}

type ConnPair struct {
	c *httputil.ServerConn
	s *httputil.ClientConn
}

type OnceRequest struct {
	cp  *ConnPair
	req *http.Request
}

type Proxy struct {
	l       net.Listener
	cfgfile string
	cfg     *JsonConfig
	mode    string
	balance string
	Pool    Container
}

func NewProxy(filename string) *Proxy {
	return &Proxy{
		cfgfile: filename,
		cfg:     NewJsonConfig(),
	}
}

func (p *Proxy) Start() error {
	if err := p.cfg.ReadCfgFile(p.cfgfile); err != nil {
		return err
	}
	fmt.Printf("%v\n", *p.cfg)
	backends, err := p.cfg.GetStringsOf("backends")
	if err != nil {
		return err
	}
	if err = p.setUpBackends(backends); err != nil {
		return err
	}

	p.mode, err = p.cfg.GetStringOf("mode")
	if err != nil {
		return err
	}

	addr, err := p.cfg.GetStringOf("listen")
	if err != nil {
		return err
	}
	l, err := p.Listen(addr)
	if err != nil {
		return err
	}

	for {
		c, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(1 * time.Second)
				continue
			}
			log.Fatalln(err)
		}
		log.Printf("[proxy] client %s connect to %s\n", c.RemoteAddr().String(), c.LocalAddr().String())
		one_req := make(chan *OnceRequest, 1)
		go p.backend_loop(one_req)
		go p.frontend_loop(c, one_req)
	}
	return err
}

func (p *Proxy) setUpBackends(backends []string) error {
	balance, err := p.cfg.GetStringOf("balance")
	if err != nil {
		return err
	}
	switch balance {
	case "rr":
		p.Pool = NewRRContainer()
	default:
		log.Fatalf("[proxy] not found banlace algorithm %s\n", balance)
	}
	p.balance = balance
	for _, addr := range backends {
		p.Pool.Put(&Server{addr: addr})
	}
	return nil
}

func (p *Proxy) Listen(addr string) (net.Listener, error) {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return l, err
}
