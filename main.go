package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

// ConvertRequest 转换请求的结构体
type ConvertRequest struct {
	HTML           string `json:"html"`
	HTMLBase64     string `json:"html_base64,omitempty"`      // Base64 编码的 HTML（可选）
	UseLocalServer bool   `json:"use_local_server,omitempty"` // 是否使用本地服务器方式
}

// ConvertResponse 转换响应的结构体
type ConvertResponse struct {
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`  // base64 编码的图片
	Error   string `json:"error,omitempty"` // 错误信息
}

var chromePool *ChromePool

func main() {
	// Initialize the Chrome instance pool on startup.
	var err error
	chromePool, err = NewChromePool(getPoolSizeFromEnv())
	if err != nil {
		log.Fatalf("Failed to create Chrome pool: %v", err)
	}
	// Setup graceful shutdown.
	defer chromePool.Shutdown()

	// 设置服务器端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 设置路由
	http.HandleFunc("/convert", handleConvert)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/", handleRoot)

	// 启动服务器
	log.Printf("HTML2Image 服务启动在端口 %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// handleRoot 处理根路径请求
func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service":   "HTML2Image API",
		"version":   "1.0.0",
		"endpoints": "POST /convert - 将 HTML 转换为图片",
	})
}

// handleHealth 健康检查端点
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"healthy": true})
}

// handleConvert 处理转换请求
func handleConvert(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只接受 POST 请求
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ConvertResponse{
			Success: false,
			Error:   "只支持 POST 请求",
		})
		return
	}

	// 解析请求体
	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ConvertResponse{
			Success: false,
			Error:   "无效的请求格式: " + err.Error(),
		})
		return
	}

	// --- Pool Usage ---
	// 1. Get an instance from the pool.
	browserCtx := chromePool.Get()
	// 2. Ensure it's returned to the pool when the function exits.
	defer chromePool.Return(browserCtx)

	// 处理 HTML 内容
	var htmlContent string

	// 优先使用 base64 编码的 HTML（如果提供）
	if req.HTMLBase64 != "" {
		decodedHTML, err := base64.StdEncoding.DecodeString(req.HTMLBase64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ConvertResponse{
				Success: false,
				Error:   "无效的 base64 编码: " + err.Error(),
			})
			return
		}
		htmlContent = string(decodedHTML)
		log.Println("使用 base64 解码的 HTML 内容")
	} else if req.HTML != "" {
		htmlContent = req.HTML
		log.Println("使用原始 HTML 内容")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ConvertResponse{
			Success: false,
			Error:   "HTML 内容不能为空（请提供 'html' 或 'html_base64' 字段）",
		})
		return
	}

	// --- Call the refactored functions ---
	var imgBytes []byte
	var err error
	if req.UseLocalServer {
		log.Println("使用本地服务器方式转换...")
		imgBytes, err = HTMLToImageWithServer(browserCtx, htmlContent)
	} else {
		log.Println("使用 data URL 方式转换...")
		imgBytes, err = HTMLToImage(browserCtx, htmlContent)

		if err != nil {
			log.Printf("Data URL 方式失败: %v，尝试本地服务器方式...", err)
			imgBytes, err = HTMLToImageWithServer(browserCtx, htmlContent)
		}
	}

	// 处理转换结果
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ConvertResponse{
			Success: false,
			Error:   "转换失败: " + err.Error(),
		})
		return
	}

	// 将图片转换为 base64
	base64Image := base64.StdEncoding.EncodeToString(imgBytes)

	// 返回成功响应
	json.NewEncoder(w).Encode(ConvertResponse{
		Success: true,
		Data:    base64Image,
	})

	log.Printf("成功转换 HTML 到图片，大小: %d 字节", len(imgBytes))
}

// HTMLToImage now accepts a browser context from the pool.
func HTMLToImage(browserCtx context.Context, htmlContent string) ([]byte, error) {
	start := time.Now()
	// Create a new context for this specific task with its own timeout.
	taskCtx, cancel := context.WithTimeout(browserCtx, 30*time.Second)
	defer cancel()

	// No longer creating allocators or new contexts here. They are managed by the pool.

	var buf []byte
	dataURL := "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString([]byte(htmlContent))

	tasks := chromedp.Tasks{
		// Note: The Navigate action will create a new tab in the shared browser instance.
		chromedp.Navigate(dataURL),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Poll(`
			document.readyState === 'complete' &&
			Array.from(document.querySelectorAll('img')).every(img => img.complete && img.naturalHeight > 0)
		`, nil, chromedp.WithPollingTimeout(15*time.Second)),
		chromedp.FullScreenshot(&buf, 100),
	}

	// Run the tasks against the task-specific context.
	if err := chromedp.Run(taskCtx, tasks); err != nil {
		return nil, fmt.Errorf("failed to run chromedp tasks: %w", err)
	}

	log.Printf("HTMLToImage 耗时: %v", time.Since(start))
	return buf, nil
}

// HTMLToImageWithServer now accepts a browser context from the pool.
func HTMLToImageWithServer(browserCtx context.Context, htmlContent string) ([]byte, error) {
	start := time.Now()
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(htmlContent))
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait a bit for the server to start and check for errors
	select {
	case <-time.After(100 * time.Millisecond):
		// server started
	case err := <-serverErr:
		return nil, fmt.Errorf("failed to start HTTP server: %w", err)
	}
	defer server.Shutdown(context.Background())

	// Create a new context for this specific task with its own timeout.
	taskCtx, cancel := context.WithTimeout(browserCtx, 30*time.Second)
	defer cancel()

	// No longer creating allocators or new contexts here.

	var buf []byte
	url := fmt.Sprintf("http://localhost:%d", port)

	tasks := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Poll(`
			document.readyState === 'complete' &&
			Array.from(document.querySelectorAll('img')).every(img => img.complete && img.naturalHeight > 0)
		`, nil, chromedp.WithPollingTimeout(15*time.Second)),
		chromedp.FullScreenshot(&buf, 90),
	}

	if err := chromedp.Run(taskCtx, tasks); err != nil {
		return nil, fmt.Errorf("failed to run chromedp tasks: %w", err)
	}

	log.Printf("HTMLToImageWithServer 耗时: %v", time.Since(start))
	return buf, nil
}
