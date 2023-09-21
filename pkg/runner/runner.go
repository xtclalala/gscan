package runner

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func NewRunner(options Options) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		options:         options,
		isRunning:       false,
		Ctx:             ctx,
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
	Ctx    context.Context
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
func (s *Runner) Open() error {
	var (
		handle  *pcap.Handle
		source  *gopacket.PacketSource
		options Options
		bpf     string
	)
	options = s.options
	handle, s.Err = pcap.OpenLive(options.Device(), options.Snapshot(), options.Promisc(), options.HandleTimeout())
	if s.Err != nil {
		return s.Err
	}
	bpf = options.Bpf()
	if bpf != "" {
		s.Err = handle.SetBPFFilter(bpf)
		if s.Err != nil {
			return s.Err
		}
	}

	source = gopacket.NewPacketSource(handle, handle.LinkType())
	s.receiveCh = source.Packets()
	s.handle = handle
	s.sendCh = make(chan []byte)
	s.isRunning = true
	return s.Err
}

// Close 关闭 chan 和 pcap.Handle
func (s *Runner) Close() {
	s.isRunning = false
	s.handle.Close()
	close(s.sendCh)
}

// DoneCh Done chan
func (s *Runner) DoneCh() {
	s.cancel()
}

// RunSender 发包
func (s *Runner) RunSender() {

	for buffer := range s.sendCh {
		s.Err = s.handle.WritePacketData(buffer)
		if s.Err != nil {
			return
		}
	}
}

// RunReceive 接收包
func (s *Runner) RunReceive() {
	if !s.isRunning {
		return
	}
	for r := range s.receiveCh {
		go s.receiveAfter(r)
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
