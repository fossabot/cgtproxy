// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package core

import (
	"github.com/black-desk/cgtproxy/internal/core/monitor"
	"github.com/black-desk/cgtproxy/internal/core/repeater"
	"github.com/black-desk/cgtproxy/internal/core/rulemanager"
	"github.com/black-desk/cgtproxy/internal/core/watcher"
)

// Injectors from wire.go:

func injectedMonitor(core *Core) (*monitor.Monitor, error) {
	v := provideOutputChan()
	config, err := provideConfig(core)
	if err != nil {
		return nil, err
	}
	cgroupRoot := provideCgroupRoot(config)
	watcher, err := provideWatcher(cgroupRoot)
	if err != nil {
		return nil, err
	}
	monitorMonitor, err := provideMonitor(v, watcher, cgroupRoot)
	if err != nil {
		return nil, err
	}
	return monitorMonitor, nil
}

func injectedRuleManager(core *Core) (*rulemanager.RuleManager, error) {
	conn, err := provideNftConn()
	if err != nil {
		return nil, err
	}
	config, err := provideConfig(core)
	if err != nil {
		return nil, err
	}
	cgroupRoot := provideCgroupRoot(config)
	bypass := provideBypass(config)
	table, err := provideTable(conn, cgroupRoot, bypass)
	if err != nil {
		return nil, err
	}
	v := provideInputChan()
	ruleManager, err := provideRuleManager(table, config, v)
	if err != nil {
		return nil, err
	}
	return ruleManager, nil
}

func injectedRepeater(core *Core) (*repeater.Repeater, error) {
	repeaterRepeater, err := provideRepeater()
	if err != nil {
		return nil, err
	}
	return repeaterRepeater, nil
}

func injectedWatcher(core *Core) (*watcher.Watcher, error) {
	config, err := provideConfig(core)
	if err != nil {
		return nil, err
	}
	cgroupRoot := provideCgroupRoot(config)
	watcherWatcher, err := provideWatcher(cgroupRoot)
	if err != nil {
		return nil, err
	}
	return watcherWatcher, nil
}
