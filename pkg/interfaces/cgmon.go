// Code generated by interfacer; DO NOT EDIT

package interfaces

import (
	"context"
	"github.com/black-desk/cgtproxy/pkg/types"
)

// CGroupMonitor is an interface generated for "github.com/black-desk/cgtproxy/pkg/cgfsmon.CGroupFSMonitor".
type CGroupMonitor interface {
	Events() <-chan types.CGroupEvents
	RunCGroupMonitor(context.Context) error
}
