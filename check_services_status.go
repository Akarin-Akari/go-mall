package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func checkPort(port string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:" + port)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}

func checkProcessByPort(port string) string {
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return "无法检查进程"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":"+port) && strings.Contains(line, "LISTENING") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				return "PID: " + fields[4]
			}
		}
	}
	return "未找到监听进程"
}

func main() {
	fmt.Println("🔍 检查前后端服务状态")
	fmt.Println("====================================================")

	// 检查后端服务 (Mall-Go)
	fmt.Println("📡 后端服务检查 (Port 8080):")
	if checkPort("8080") {
		fmt.Println("  ✅ 后端服务正在运行")
		fmt.Println("  📍 进程信息:", checkProcessByPort("8080"))

		// 测试健康检查端点
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://localhost:8080/health")
		if err != nil {
			fmt.Println("  ⚠️  健康检查失败:", err)
		} else {
			fmt.Printf("  ✅ 健康检查通过 (状态码: %d)\n", resp.StatusCode)
			resp.Body.Close()
		}
	} else {
		fmt.Println("  ❌ 后端服务未运行")
		fmt.Println("  💡 启动命令: cd mall-go && go run cmd/server/main.go")
	}

	fmt.Println()

	// 检查前端服务 (Next.js)
	fmt.Println("🌐 前端服务检查:")
	frontendPorts := []string{"3000", "3001", "5173", "8081"}
	frontendRunning := false

	for _, port := range frontendPorts {
		if checkPort(port) {
			fmt.Printf("  ✅ 前端服务正在端口 %s 运行\n", port)
			fmt.Printf("  📍 进程信息: %s\n", checkProcessByPort(port))
			frontendRunning = true
			break
		}
	}

	if !frontendRunning {
		fmt.Println("  ❌ 前端服务未运行")
		fmt.Println("  💡 启动命令: cd mall-frontend && npm run dev")
	}

	fmt.Println()
	fmt.Println("====================================================")

	// 总结
	backendOK := checkPort("8080")
	frontendOK := false
	for _, port := range frontendPorts {
		if checkPort(port) {
			frontendOK = true
			break
		}
	}

	fmt.Println("📊 服务状态总结:")
	if backendOK && frontendOK {
		fmt.Println("  🎉 前后端服务都在正常运行，可以进行联调测试！")
	} else if backendOK {
		fmt.Println("  ⚠️  后端服务正常，但前端服务需要启动")
	} else if frontendOK {
		fmt.Println("  ⚠️  前端服务正常，但后端服务需要启动")
	} else {
		fmt.Println("  ❌ 前后端服务都需要启动")
	}
}
