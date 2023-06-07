package rulemanager

import (
	"errors"
	"net"
	"os"

	"github.com/black-desk/cgtproxy/internal/config"
	"github.com/black-desk/cgtproxy/internal/core/table"
	. "github.com/black-desk/cgtproxy/internal/log"
	"github.com/black-desk/cgtproxy/internal/types"
	. "github.com/black-desk/lib/go/errwrap"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func (m *RuleManager) Run() (err error) {
	defer Wrap(&err, "Error occurs while running the nftable rules manager.")

	defer m.removeRoute()
	err = m.addRoute()
	if err != nil {
		return
	}

	defer m.removeNftableRules()
	err = m.initializeNftableRuels()
	if err != nil {
		return
	}

	for event := range m.cgroupEventChan {
		switch event.EventType {
		case types.CgroupEventTypeNew:
			m.handleNewCgroup(event.Path)
		case types.CgroupEventTypeDelete:
			m.handleDeleteCgroup(event.Path)
		}
	}
	return
}

func (m *RuleManager) initializeNftableRuels() (err error) {
	defer Wrap(&err, "Failed to initialize nftable ruels.")

	for _, tp := range m.cfg.TProxies {
		err = m.nft.AddChainAndRulesForTProxy(tp)
		if err != nil {
			return
		}

		err = m.addRule(tp.Mark)
		if err != nil {
			return
		}
	}

	return
}

func (m *RuleManager) removeNftableRules() {
	err := m.nft.Clear()
	if err != nil {
		Log.Errorw("Failed to delete nft table.",
			"error", err,
		)
	}

	for _, rule := range m.rule {
		err = netlink.RuleDel(rule)
		if err == nil {
			continue
		}

		Log.Errorw("Failed to delete route rule.",
			"rule", rule,
			"error", err,
		)
	}
	return
}

func (m *RuleManager) addRule(mark config.RerouteMark) (err error) {
	defer Wrap(&err, "Failed to add route rule.")

	Log.Infow("Adding route rule.",
		"mark", mark,
		"table", m.cfg.RouteTable,
	)

	// ip rule add fwmark <mark> lookup <table>

	rule := netlink.NewRule()
	rule.Family = netlink.FAMILY_ALL
	rule.Mark = int(mark) // WARN(black_desk): ???
	rule.Table = m.cfg.RouteTable

	err = netlink.RuleAdd(rule)
	if errors.Is(err, os.ErrExist) {
		Log.Infow("Rule already exists.")
		err = nil
	}
	if err != nil {
		return
	}

	m.rule = append(m.rule, rule)

	return
}

func (m *RuleManager) addRoute() (err error) {
	defer Wrap(&err, "Failed to add route.")

	Log.Infow("Adding route.",
		"table", m.cfg.RouteTable,
	)

	// ip route add local default dev lo table <table>

	var iface *net.Interface
	iface, err = net.InterfaceByName("lo")
	if err != nil {
		return
	}

	cidrStrs := []string{"0.0.0.0/0", "0::0/0"}

	for _, cidrStr := range cidrStrs {
		var cidr *net.IPNet

		_, cidr, err = net.ParseCIDR(cidrStr)
		if err != nil {
			return
		}

		route := &netlink.Route{
			LinkIndex: iface.Index,
			Scope:     unix.RT_SCOPE_HOST,
			Dst:       cidr,
			Table:     m.cfg.RouteTable,
			Type:      unix.RTN_LOCAL,
		}

		err = netlink.RouteAdd(route)
		if err != nil {
			return
		}

		m.route = append(m.route, route)
	}

	return
}

func (m *RuleManager) removeRoute() {
	for i := range m.route {
		err := netlink.RouteDel(m.route[i])

		if err == nil {
			continue
		}

		Log.Warnw("Failed to remove route",
			"error", err)
	}

	return
}

func (m *RuleManager) handleNewCgroup(path string) {
	var target table.Target
	for i := range m.matchers {
		if !m.matchers[i].reg.Match([]byte(path)) {
			continue
		}

		Log.Infow("Rule found for this cgroup",
			"cgroup", path,
			"rule", m.cfg.Rules[i].String(),
		)

		target = m.matchers[i].target

		break
	}

	if target.Op == table.TargetNoop {
		Log.Debugw("No rule match this cgroup",
			"cgroup", path,
		)
		return
	}

	err := m.nft.AddCgroup(path, &target)
	if err != nil {
		Log.Errorw("Failed to update nft for new cgroup",
			"error", err,
		)
	}
}

func (m *RuleManager) handleDeleteCgroup(path string) {
	err := m.nft.RemoveCgroup(path)
	if err != nil {
		Log.Errorw("Failed to update nft for removed cgroup", "error", err)
	}
}
