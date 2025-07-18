package printer

import (
	"fmt"
	"net"
	"time"

	"github.com/xiaohao0576/odoo-epos/raster"
	"github.com/xiaohao0576/odoo-epos/transformer"
)

type TCPPrinter struct {
	paperWidth        int                         // 纸张宽度
	marginBottom      int                         // 下边距
	cutCommand        []byte                      //切纸命令
	cashDrawerCommand []byte                      // 钱箱命令
	HostPort          string                      // 打印机地址
	fd                net.Conn                    // 直接用 net.Conn
	transformer       transformer.TransformerFunc // 用于转换图像的转换器
}

func (p *TCPPrinter) String() string {
	return fmt.Sprintf("TCPPrinter{HostPort: %s, paperWidth: %d, marginBottom: %d}", p.HostPort, p.paperWidth, p.marginBottom)
}

func (p *TCPPrinter) Open() error {
	if p.HostPort == "" {
		return net.ErrClosed
	}
	if p.fd != nil {
		p.fd.Close()
	}
	conn, err := net.DialTimeout("tcp", p.HostPort, 5*time.Second)
	if err != nil {
		return err
	}
	p.fd = conn
	return nil
}

func (p *TCPPrinter) Close() error {
	if p.fd != nil {
		err := p.fd.Close()
		p.fd = nil
		return err
	}
	return nil
}

func (p *TCPPrinter) OpenCashBox() error {
	if p.fd == nil {
		if err := p.Open(); err != nil {
			return err
		}
	}
	defer p.Close() // 确保在函数结束时关闭连接
	// 发送打开钱箱的命令
	_, err := p.fd.Write(p.cashDrawerCommand)
	return err
}

func (p *TCPPrinter) PrintRasterImage(img *raster.RasterImage) error {
	img = p.transformer(img) // 使用转换器转换图像
	if img == nil {
		return nil // 如果转换器返回 nil，表示不需要打印图像
	}
	if p.fd == nil {
		if err := p.Open(); err != nil {
			return err
		}
	}
	defer p.Close()
	for _, page := range img.CutPages() {
		page.AutoMarginLeft(p.paperWidth)
		page.AddMarginBottom(p.marginBottom)
		if _, err := p.fd.Write(page.ToEscPosRasterCommand(1024)); err != nil {
			return err
		}
		if _, err := p.fd.Write(p.cutCommand); err != nil {
			return err
		}
	}

	return nil
}

func (p *TCPPrinter) PrintRaw(data []byte) error {
	if p.fd == nil {
		if err := p.Open(); err != nil {
			return err
		}
	}
	defer p.Close()
	if len(data) == 0 {
		return fmt.Errorf("no data to print")
	}
	if _, err := p.fd.Write(data); err != nil {
		return fmt.Errorf("failed to write data to printer: %w", err)
	}
	return nil
}
