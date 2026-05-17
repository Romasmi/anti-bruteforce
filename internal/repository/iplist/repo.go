package iplist

import (
	"database/sql"
	"net"

	"github.com/Romasmi/anti-bruteforce/pkg/ratelimiter/algorithms/leackybucket"
)

var _ leackybucket.Repository = (*IpRepo)(nil)

type IpRepo struct {
	db *sql.DB
}

func NewIpRepo(db *sql.DB) *IpRepo {
	return &IpRepo{db: db}
}

func (r *IpRepo) AddToWhitelist(cidr string) error {
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`INSERT INTO ip_whitelist (network) VALUES ($1) ON CONFLICT DO NOTHING`, n.String())
	return err
}

func (r *IpRepo) RemoveFromWhitelist(cidr string) error {
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM ip_whitelist WHERE network = $1`, n.String())
	return err
}

func (r *IpRepo) AddToBlacklist(cidr string) error {
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`INSERT INTO ip_blacklist (network) VALUES ($1) ON CONFLICT DO NOTHING`, n.String())
	return err
}

func (r *IpRepo) RemoveFromBlacklist(cidr string) error {
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM ip_blacklist WHERE network = $1`, n.String())
	return err
}

func (r *IpRepo) ExistsInWhiteList(ip string) bool {
	return r.existsIn("ip_whitelist", ip)
}

func (r *IpRepo) ExistsInBlackList(ip string) bool {
	return r.existsIn("ip_blacklist", ip)
}

func (r *IpRepo) existsIn(table, ip string) bool {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM `+table+` WHERE network >> $1::inet)`,
		ip,
	).Scan(&exists)
	return err == nil && exists
}
