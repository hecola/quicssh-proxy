package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	//"io"
	"log"
	"math/big"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	quic "github.com/quic-go/quic-go"
	cli "github.com/urfave/cli/v2"
)

func server(c *cli.Context) error {
	// 创建带取消功能的上下文，以便进行优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 捕获系统信号以便优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("接收到信号: %v，正在关闭服务器...", sig)
		cancel()
	}()

	// 配置 TLS
	certPath := c.String("cert")
	keyPath := c.String("key")
	tlsConfig, err := loadTLSConfig(certPath, keyPath)
	if err != nil {
		log.Printf("加载 TLS 配置错误: %v", err)
		return err
	}

	// 配置 QUIC
	quicConfig := &quic.Config{
		KeepAlivePeriod: 30 * time.Second,
		MaxIdleTimeout:  300 * time.Second,
		// MaxIncomingStreams: 1000, // 根据需求调整
	}

	// 配置监听地址
	bindAddr := c.String("bind")
	listener, err := quic.ListenAddr(bindAddr, tlsConfig, quicConfig)
	if err != nil {
		log.Fatalf("无法监听地址 %q: %v", bindAddr, err)
	}
	defer listener.Close()
	log.Printf("服务器正在监听 %q...", bindAddr)

	var wg sync.WaitGroup

	for {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			log.Println("服务器关闭中...")
			wg.Wait()
			return nil
		default:
			// 使用 Accept 方法，阻塞等待新会话
			session, err := listener.Accept(ctx)
			if err != nil {
				// 如果上下文被取消，退出循环
				if ctx.Err() != nil {
					return nil
				}
				log.Printf("接受连接错误: %v", err)
				continue
			}

			log.Printf("接受到新的会话来自 %s", session.RemoteAddr())
			wg.Add(1)
			go func(sess quic.Connection) {
				defer wg.Done()
				serverSessionHandler(ctx, sess)
			}(session)
		}
	}
}

func loadTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	// 检查证书和密钥文件是否存在
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		log.Println("TLS 证书或密钥不存在，正在生成自签名证书...")
		if err := generateSelfSignedCert(certPath, keyPath); err != nil {
			return nil, err
		}
	}

	// 加载证书和密钥
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quicssh"},
		MinVersion:   tls.VersionTLS12, // 设置最低 TLS 版本
	}

	return tlsConfig, nil
}

func generateSelfSignedCert(certPath, keyPath string) error {
	// 生成 RSA 密钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 创建自签名证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // 1 年有效期
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 获取本地主机名和 IP 地址
	host, err := os.Hostname()
	if err != nil {
		return err
	}
	template.DNSNames = []string{host}
	template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	// 创建证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return err
	}

	// 编码并保存密钥
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return err
	}

	// 编码并保存证书
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return err
	}

	log.Printf("自签名证书已生成并保存到 %s 和 %s", certPath, keyPath)
	return nil
}

func serverSessionHandler(ctx context.Context, session quic.Connection) {
	log.Printf("处理会话来自 %s", session.RemoteAddr())
	defer func() {
		if err := session.CloseWithError(0, "关闭会话"); err != nil {
			log.Printf("关闭会话错误: %v", err)
		}
	}()

	for {
		// 使用 AcceptStream 来接受新的流
		stream, err := session.AcceptStream(ctx)
		if err != nil {
			// 如果上下文被取消，退出
			if ctx.Err() != nil {
				return
			}
			log.Printf("接受流错误: %v", err)
			return
		}

		log.Printf("接受到新的流来自 %s", session.RemoteAddr())
		go serverStreamHandler(ctx, session, stream)
	}
}

func serverStreamHandler(ctx context.Context, session quic.Connection, stream quic.Stream) {
	log.Printf("处理来自 %s 的流", session.RemoteAddr())
	defer func() {
		if err := stream.Close(); err != nil {
			log.Printf("关闭流错误: %v", err)
		}
	}()

	// 连接到本地 SSH 服务器
	sshAddr := "127.0.0.1:22"
	rConn, err := net.Dial("tcp", sshAddr)
	if err != nil {
		log.Printf("连接到 SSH 服务器 %s 失败: %v", sshAddr, err)
		return
	}
	defer rConn.Close()

	// 创建子上下文，以便在任一方向出错时取消另一个方向
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	// 从流读取并写入到 SSH 连接
	go func() {
		defer wg.Done()
		if err := ReadAndWrite(streamCtx, stream, rConn); err != nil {
			log.Printf("流到 SSH 连接错误: %v", err)
			cancel()
		}
	}()

	// 从 SSH 连接读取并写入到流
	go func() {
		defer wg.Done()
		if err := ReadAndWrite(streamCtx, rConn, stream); err != nil {
			log.Printf("SSH 连接到流错误: %v", err)
			cancel()
		}
	}()

	// 等待任一方向完成
	wg.Wait()
	log.Printf("完成流 %s 的数据传输", session.RemoteAddr())
}
