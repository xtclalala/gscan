package runner

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/xtclalala/ylog"
)

func NewRunner(options Options) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		options:         options,
		isRunning:       false,
		ctx:             ctx,
		cancel:          cancel,
		parseHandleFunc: make([]ParseHandle, 0, 5),
	}
}

type Runner struct {
	// 操作网卡
	handle *pcap.Handle
	// 发包数据 chan
	sendCh chan []byte
	// 接收包 chan
	receiveCh chan gopacket.Packet
	// 相关设置
	options Options
	// flags
	isRunning bool
	// 控制结束
	ctx    context.Context
	cancel context.CancelFunc
	// 包接收后的操作
	parseHandleFunc []ParseHandle

	Err error
}

type ParseHandle func(gopacket.Packet) bool

// AppendParseHandle 添加接收报的解析方法
func (s *Runner) AppendParseHandle(f ParseHandle) {
	var (
		handles []ParseHandle
	)
	handles = s.parseHandleFunc

	handles = append(handles, f)
	s.parseHandleFunc = handles
}

func (s *Runner) PushPacket(packet []byte) {
	s.sendCh <- packet
}

// Open 启动
func (s *Runner) Open() (err error) {
	var (
		handle  *pcap.Handle
		source  *gopacket.PacketSource
		options Options
		bpf     string
	)
	options = s.options
	handle, err = pcap.OpenLive(options.Device(), options.Snapshot(), options.Promisc(), options.HandleTimeout())
	if err != nil {
		return
	}
	bpf = options.Bpf()
	if bpf != "" {
		err = handle.SetBPFFilter(bpf)
		if err != nil {
			ylog.WithFields(map[string]string{
				"command": "runner",
				"BPF":     bpf + "is failed",
			}).Errorf(err.Error())
			return
		}
	}

	source = gopacket.NewPacketSource(handle, handle.LinkType())
	s.receiveCh = source.Packets()
	s.handle = handle
	s.sendCh = make(chan []byte)
	s.isRunning = true
	s.Err = err
	return
}

// Close 关闭 chan 和 pcap.Handle
func (s *Runner) Close() {
	s.isRunning = false
	s.handle.Close()
	close(s.sendCh)
	return
}

// DoneCh Done chan
func (s *Runner) DoneCh() {
	s.ctx.Done()
}

// RunSender 发包
func (s *Runner) RunSender() {
	var err error

	for {
		select {
		case buffer := <-s.sendCh:
			err = s.handle.WritePacketData(buffer)
			if err != nil {
				ylog.WithField("command", "runner").Errorf(err.Error())
			}
		case <-s.ctx.Done():
			ylog.WithField("command", "runner").Infof("sender is done")
			return
		}
	}

}

// RunReceive 接收包
func (s *Runner) RunReceive() {
	if !s.isRunning {
		return
	}
	for {
		select {
		case r := <-s.receiveCh:
			go s.receiveAfter(r)
		case <-s.ctx.Done():
			ylog.WithField("command", "runner").Infof("receive is done")
			return
		}
	}
}

// 传入解析包,执行解析操作
func (s *Runner) receiveAfter(p gopacket.Packet) {
	if !s.isRunning {
		return
	}
	var result bool
	for _, f := range s.parseHandleFunc {
		result = f(p)
		if result {
			break
		}
	}
	//s.wg.Done()
}
