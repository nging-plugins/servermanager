package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)
	go func() {
		for sig := range signalChan {
			fmt.Printf("[%s] 收到信号: %v\n", time.Now().Format(time.DateTime), sig)
			// 如果需要特定处理，可以在这里添加逻辑
			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				fmt.Println("优雅退出...")
				os.Exit(0)
			}
		}
	}()

	fmt.Println("[" + time.Now().Format(time.DateTime) + "] 程序运行中，按 Ctrl+C 退出...")
	select {} // 阻塞主线程
}
