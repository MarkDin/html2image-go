package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
)

// ChromePool manages a pool of reusable Chrome browser instances.
type ChromePool struct {
	allocator context.Context
	cancel    context.CancelFunc
	pool      chan context.Context
}

// NewChromePool creates and initializes a new pool of Chrome instances.
func NewChromePool(maxInstances int) (*ChromePool, error) {
	if maxInstances <= 0 {
		maxInstances = 5 // Default to 5 instances
	}
	log.Printf("Initializing Chrome pool with %d instances...", maxInstances)

	// --- Create a shared allocator ---
	// All browser instances in the pool will be created from this single allocator.
	opts := getOptimizedChromeOptions()
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	p := &ChromePool{
		allocator: allocCtx,
		cancel:    cancel,
		pool:      make(chan context.Context, maxInstances),
	}

	// --- Create and add instances to the pool ---
	for i := 0; i < maxInstances; i++ {
		ctx, _ := chromedp.NewContext(p.allocator)

		// Perform a "warm-up" navigation to initialize the browser instance.
		if err := chromedp.Run(ctx, chromedp.Navigate("about:blank")); err != nil {
			log.Printf("Warning: Failed to warm up Chrome instance %d: %v", i+1, err)
			// Still add it to the pool, it might recover.
		}
		p.pool <- ctx
	}

	log.Println("Chrome pool initialized successfully.")
	return p, nil
}

// Get borrows a Chrome instance from the pool.
// It will block until an instance is available.
func (p *ChromePool) Get() context.Context {
	return <-p.pool
}

// Return gives a Chrome instance back to the pool.
func (p *ChromePool) Return(ctx context.Context) {
	// In a more complex scenario, you might check if the context is still valid
	// before returning it to the pool. For now, we'll just return it.
	p.pool <- ctx
}

// Shutdown gracefully closes all browser instances and the allocator.
func (p *ChromePool) Shutdown() {
	log.Println("Shutting down Chrome pool...")

	// 设置 60 秒超时
	timeout := time.After(60 * time.Second)
	closed := 0

	// Close all browser contexts
	for i := 0; i < cap(p.pool); i++ {
		select {
		case ctx := <-p.pool:
			chromedp.Cancel(ctx)
			closed++
		case <-timeout:
			log.Printf("Warning: Shutdown timeout after 60s, only closed %d/%d contexts", closed, cap(p.pool))
			goto cleanup
		}
	}

cleanup:
	// Close the allocator context
	p.cancel()
	log.Println("Chrome pool shut down.")
}

// getOptimizedChromeOptions 返回一组用于提升性能的 Chrome 启动参数
func getOptimizedChromeOptions() []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		// --- 无头模式，确保浏览器在后台运行 ---
		chromedp.Flag("headless", true),

		// --- 安全性和稳定性 ---
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),

		// --- 性能优化 ---
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-component-update", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),

		// --- 与内容加载相关 ---
		chromedp.Flag("allow-running-insecure-content", true),
	}
}

// getPoolSizeFromEnv reads the pool size from an environment variable.
func getPoolSizeFromEnv() int {
	sizeStr := os.Getenv("CHROME_POOL_SIZE")
	if sizeStr == "" {
		return 5 // Default size
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		log.Printf("Invalid CHROME_POOL_SIZE '%s', using default of 5.", sizeStr)
		return 5
	}
	return size
}
