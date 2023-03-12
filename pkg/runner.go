package pkg

import (
	"github.com/xtclalala/infoK1t/pkg/runner"
	"sync"
)

var (
	r    *runner.Runner
	once sync.Once
)

func DefaultRunner() *runner.Runner {
	once.Do(func() {
		d = DefaultDevice()
		option := runner.NewOptions(d.PcapName, 1024, false, 10, "!src "+d.Ipv4.String())
		r = runner.NewRunner(*option)
		r.Open()
	})

	return r
}
