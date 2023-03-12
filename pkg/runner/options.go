package runner

import "time"

func NewOptions(device string, snapshot int32, promisc bool, handleTimeout time.Duration, bpf string) *Options {
	var o = &Options{}
	o.SetDevice(device)
	o.SetSnapshot(snapshot)
	o.SetPromisc(promisc)
	o.SetHandleTimeout(handleTimeout)
	o.SetBpf(bpf)
	return o
}

type Options struct {
	// 网卡名称
	device string
	// 设置接收的包大小 1024 - 65535
	snapshot int32
	// 混杂模式 true -> 接收所有的额包  false -> 只接收自己的包
	promisc bool
	// 网卡
	handleTimeout time.Duration
	// 网卡过滤器
	bpf string
}

func (o *Options) Device() string {
	return o.device
}

func (o *Options) SetDevice(device string) {
	o.device = device
}

func (o *Options) Snapshot() int32 {
	return o.snapshot
}

func (o *Options) SetSnapshot(snapshot int32) {
	o.snapshot = snapshot
}

func (o *Options) Promisc() bool {
	return o.promisc
}

func (o *Options) SetPromisc(promisc bool) {
	o.promisc = promisc
}

func (o *Options) HandleTimeout() time.Duration {
	return o.handleTimeout
}

func (o *Options) SetHandleTimeout(handleTimeout time.Duration) {
	o.handleTimeout = handleTimeout
}

func (o *Options) Bpf() string {
	return o.bpf
}

func (o *Options) SetBpf(bpf string) {
	o.bpf = bpf
}
