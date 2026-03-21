package dns

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	pq "github.com/lib/pq"
	mdns "github.com/miekg/dns"
	"github.com/openfiltr/openfiltr/internal/config"
	"github.com/openfiltr/openfiltr/internal/storage"
)

type Server struct {
	cfg        *config.Config
	db         *sql.DB
	server     *mdns.Server
	regexMu    sync.RWMutex
	regexCache map[string]cachedRegexp
}

type cachedRegexp struct {
	re  *regexp.Regexp
	err error
}

func NewServer(cfg *config.Config, db *sql.DB) *Server {
	return &Server{cfg: cfg, db: db, regexCache: make(map[string]cachedRegexp)}
}

func (s *Server) Start() error {
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
		domain := normaliseDomain(q.Name)
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
	domain = normaliseDomain(domain)
	if domain == "" {
		return false
	}

	if s.hasExactBlockRule(domain) {
		return true
	}
	if s.hasWildcardBlockRule(domain) {
		return true
	}
	return s.hasRegexBlockRule(domain)
}

func (s *Server) hasExactBlockRule(domain string) bool {
	var n int
	if err := s.db.QueryRow(storage.Rebind(`SELECT COUNT(*) FROM block_rules WHERE enabled=1 AND rule_type='exact' AND lower(pattern)=?`), domain).Scan(&n); err != nil {
		return false
	}
	return n > 0
}

func (s *Server) hasWildcardBlockRule(domain string) bool {
	patterns := wildcardPatterns(domain)
	if len(patterns) == 0 {
		return false
	}

	var n int
	if err := s.db.QueryRow(storage.Rebind(`SELECT COUNT(*) FROM block_rules WHERE enabled=1 AND rule_type='wildcard' AND lower(pattern) = ANY(?)`), pq.Array(patterns)).Scan(&n); err != nil {
		return false
	}
	return n > 0
}

func (s *Server) hasRegexBlockRule(domain string) bool {
	rows, err := s.db.Query(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type='regex'`)
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var pattern string
		if err := rows.Scan(&pattern); err != nil {
			continue
		}
		re, err := s.compiledRegex(pattern)
		if err != nil {
			slog.Warn("invalid block rule regex", "pattern", pattern, "err", err)
			continue
		}
		if re.MatchString(domain) {
			return true
		}
	}

	return false
}

func wildcardPatterns(domain string) []string {
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return nil
	}

	patterns := make([]string, 0, len(labels)-1)
	for i := 1; i < len(labels); i++ {
		patterns = append(patterns, "*."+strings.Join(labels[i:], "."))
	}
	return patterns
}

func normaliseDomain(domain string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
}

func (s *Server) compiledRegex(pattern string) (*regexp.Regexp, error) {
	s.regexMu.RLock()
	cached, ok := s.regexCache[pattern]
	s.regexMu.RUnlock()
	if ok {
		return cached.re, cached.err
	}

	re, err := regexp.Compile("(?i)" + pattern)

	s.regexMu.Lock()
	s.regexCache[pattern] = cachedRegexp{re: re, err: err}
	s.regexMu.Unlock()

	return re, err
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
