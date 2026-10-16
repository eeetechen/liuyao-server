// Command server 启动六爻排盘 HTTP 服务。
//
// 用法：
//
//	go run ./cmd/server                 # 默认监听 :8080
//	go run ./cmd/server -addr :9000     # 指定端口
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apple/liuyao/internal/api"
)

func main() {
	addr := flag.String("addr", ":8080", "监听地址")
	flag.Parse()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           api.New(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 优雅关闭
	idle := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("正在关闭服务…")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("关闭出错: %v", err)
		}
		close(idle)
	}()

	log.Printf("六爻排盘服务已启动 → http://localhost%s", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}
	<-idle
	log.Println("服务已停止")
}
