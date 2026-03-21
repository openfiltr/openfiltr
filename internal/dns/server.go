package dns

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	mdns "github.com/miekg/dns"
	"github.com/openfiltr/openfiltr/internal/config"
	"github.com/openfiltr/openfiltr/internal/storage"
)

type Server struct {
	cfg        *config.Config
	db         *sql.DB
	server     *mdns.Server
	blockRules *blockRuleMatcher
}

func NewServer(cfg *config.Config, db *sql.DB) *Server {
	return &Server{cfg: cfg, db: db, blockRules: newBlockRuleMatcher(db)}
}

func (s *Server) Start() error {
	if err := s.blockRules.prime(); err != nil {
		return err
	}

	mux := mdns.NewServeMux()
	mux.HandleFunc(".", s.handle)
	s.server = &mdns.Server{Addr: s.cfg.Server.ListenDNS, Net: "udp", Handler: mux}
	slog.Info("DNS server listening", "addr", s.cfg.Server.ListenDNS)
	return s.server.ListenAndServe()
}

func (s *Server) Stop() {
	if s.server != nil {
		_ = s.server.Shutdown()
	}
}

func (s *Server) handle(w mdns.ResponseWriter, r *mdns.Msg) {
	start := time.Now()
	m := new(mdns.Msg)
	m.SetReply(r)
	m.Authoritative = false
	m.RecursionAvailable = true

	clientIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	for _, q := range r.Question {
		domain := strings.TrimSuffix(strings.ToLower(q.Name), ".")
		qtype := mdns.TypeToString[q.Qtype]
		action := "allowed"

		if s.isBlocked(domain) {
			action = "blocked"
			m.Rcode = mdns.RcodeNameError
		} else if rrs := s.localEntries(domain, q.Qtype); len(rrs) > 0 {
			m.Answer = append(m.Answer, rrs...)
		} else {
			if err := s.forward(r, m); err != nil {
				slog.Error("DNS forward error", "domain", domain, "err", err)
				m.Rcode = mdns.RcodeServerFailure
			}
		}

		ms := int(time.Since(start).Milliseconds())
		go s.log(clientIP, domain, qtype, action, ms)
	}

	if err := w.WriteMsg(m); err != nil {
		slog.Error("DNS write error", "err", err)
	}
}

func (s *Server) isBlocked(domain string) bool {
	if s.blockRules == nil {
		s.blockRules = newBlockRuleMatcher(s.db)
	}
	return s.blockRules.matches(domain)
}

func (s *Server) localEntries(domain string, qtype uint16) []mdns.RR {
	typeName := mdns.TypeToString[qtype]
	rows, err := s.db.Query(storage.Rebind(`SELECT entry_type,value,ttl FROM dns_entries WHERE host=? AND enabled=1 AND entry_type=?`), domain, typeName)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var rrs []mdns.RR
	for rows.Next() {
		var et, val string
		var ttl int
		if err := rows.Scan(&et, &val, &ttl); err != nil {
			continue
		}
		switch et {
		case "A":
			if ip := net.ParseIP(val).To4(); ip != nil {
				rrs = append(rrs, &mdns.A{Hdr: mdns.RR_Header{Name: domain + ".", Rrtype: mdns.TypeA, Class: mdns.ClassINET, Ttl: uint32(ttl)}, A: ip})
			}
		case "AAAA":
			if ip := net.ParseIP(val); ip != nil {
				rrs = append(rrs, &mdns.AAAA{Hdr: mdns.RR_Header{Name: domain + ".", Rrtype: mdns.TypeAAAA, Class: mdns.ClassINET, Ttl: uint32(ttl)}, AAAA: ip})
			}
		case "CNAME":
			rrs = append(rrs, &mdns.CNAME{Hdr: mdns.RR_Header{Name: domain + ".", Rrtype: mdns.TypeCNAME, Class: mdns.ClassINET, Ttl: uint32(ttl)}, Target: val + "."})
		}
	}
	return rrs
}

func (s *Server) forward(req, resp *mdns.Msg) error {
	c := &mdns.Client{Timeout: 5 * time.Second}
	for _, up := range s.cfg.DNS.UpstreamServers {
		addr := up.Address
		if !strings.Contains(addr, ":") {
			addr += ":53"
		}
		r, _, err := c.Exchange(req, addr)
		if err != nil {
			slog.Warn("upstream DNS failed", "upstream", up.Name, "err", err)
			continue
		}
		resp.Answer, resp.Ns, resp.Extra, resp.Rcode = r.Answer, r.Ns, r.Extra, r.Rcode
		return nil
	}
	return fmt.Errorf("all upstreams failed")
}

func (s *Server) log(clientIP, domain, qtype, action string, ms int) {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	id := fmt.Sprintf("%x", b)
	_, _ = s.db.Exec(storage.Rebind(`INSERT INTO activity_log(id,client_ip,domain,query_type,action,response_time_ms) VALUES(?,?,?,?,?,?)`),
		id, clientIP, domain, qtype, action, ms)
}
