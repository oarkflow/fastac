package matcher

import (
	"github.com/oarkflow/fastac/model/defs"
	"github.com/oarkflow/fastac/model/fm"
)

type IMatcher interface {
	GetPolicyKey() string
	RangeMatches(rDef defs.RequestDef, rvals []interface{}, fMap fm.FunctionMap, fn func(rule []string) bool) error
}
