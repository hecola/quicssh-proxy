// client.go
package main

import (
	"crypto/tls"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	quic "github.com/quic-go/quic-go"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/net/context"
)

func client(c *cli.Context) error {
	// 设置上下文，添加超时和取消功能
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// 捕获系统信号以便优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("接收到信号: %v，正在关闭客户端...", sig)
		cancel()
	}()

	// 配置 TLS，建议使用验证证书而非跳过验证
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // 如果有可能，请使用正确的证书验证
		NextProtos:         []string{"quicssh"},
	}

	// QUIC 配置
	quicConfig := &quic.Config{
		KeepAlivePeriod: 30 * time.Second,
		MaxIdleTimeout:  300 * time.Second,
		// MaxIncomingStreams: 1000, // 根据需求调整
	}

	addr := c.String("addr")
	log.Printf("正在连接到 %q...", addr)

	// 建立 QUIC 会话
	session, err := quic.DialAddr(ctx, addr, tlsConfig, quicConfig)
	if err != nil {
		return err
	}
	defer func() {
		if err := session.CloseWithError(0, "关闭会话"); err != nil {
			log.Printf("会话关闭错误: %v", err)
		}
	}()

	log.Println("正在打开同步流...")
	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Printf("流关闭错误: %v", err)
		}
	}()

	log.Println("开始数据传输...")

	var wg sync.WaitGroup
	wg.Add(2)

	// 错误通道
	errChan := make(chan error, 2)

	// 从流读取并写入标准输出
	go func() {
		defer wg.Done()
		if err := ReadAndWrite(ctx, stream, os.Stdout); err != nil {
			errChan <- err
		}
	}()

	// 从标准输入读取并写入流
	go func() {
		defer wg.Done()
		if err := ReadAndWrite(ctx, os.Stdin, stream); err != nil {
			errChan <- err
		}
	}()

	// 等待任意一个错误发生
	select {
	case err := <-errChan:
		if err != nil && err != io.EOF {
			log.Printf("传输错误: %v", err)
			cancel()
		}
	case <-ctx.Done():
		log.Println("上下文已取消")
	}

	wg.Wait()
	return nil
}
