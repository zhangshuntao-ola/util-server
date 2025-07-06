package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 请求结构
type TestRequest struct {
	AppID    int      `json:"app_id"`
	RoleDesc string   `json:"role_desc"`
	Scenes   []string `json:"scenes"`
	Style    string   `json:"style"`
	Callback string   `json:"callback"`
}

// 响应结构
type TestResponse struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	TaskID     string `json:"task_id"`
	CostTime   int    `json:"cost_time"`
	QueueCount int    `json:"queue_count"`
}

// 回调结构
type CallbackRequest struct {
	TaskID  string `json:"task_id"`
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    struct {
		Imgs []struct {
			URL   string `json:"url"`
			Index int    `json:"index"`
		} `json:"imgs"`
	} `json:"data"`
}

// CSV行结构
type CSVRow struct {
	AppID    int
	RoleDesc string
	Scenes   []string
	Style    string
}

// 全局变量
var (
	testFolder      string
	apiURL          = "http://192.168.11.4:7865/comfyui/scene_text2img"
	callbackURL     = "http://localhost:9983/callback"
	requestInterval = 20 * time.Second // 请求间隔时间
	retryWaitBase   = 10 * time.Second // 重试等待基础时间
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "test" {
		runTests()
		return
	}

	// 启动web服务
	startWebServer()
}

// 运行测试用例
func runTests() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run main.go test <csv_file> [interval_seconds]")
	}

	csvFile := os.Args[2]

	// 如果提供了间隔时间参数，使用它
	if len(os.Args) > 3 {
		if interval, err := strconv.Atoi(os.Args[3]); err == nil && interval > 0 {
			requestInterval = time.Duration(interval) * time.Second
			fmt.Printf("使用自定义请求间隔: %v\n", requestInterval)
		}
	}

	// 创建测试文件夹
	testFolder = fmt.Sprintf("test-%s", time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(testFolder, 0755); err != nil {
		log.Fatal("创建测试文件夹失败:", err)
	}

	fmt.Printf("创建测试文件夹: %s\n", testFolder)

	// 读取CSV文件
	rows, err := readCSV(csvFile)
	if err != nil {
		log.Fatal("读取CSV文件失败:", err)
	}

	// 发送测试请求
	for i, row := range rows {
		fmt.Printf("发送测试请求 %d/%d\n", i+1, len(rows))

		// 重试机制
		maxRetries := 3
		for retry := 0; retry < maxRetries; retry++ {
			if err := sendTestRequest(row); err != nil {
				if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
					// 如果是频率限制错误，等待更长时间再重试
					waitTime := time.Duration(10*(retry+1)) * time.Second
					log.Printf("请求频率限制，等待 %v 后重试... (重试 %d/%d)", waitTime, retry+1, maxRetries)
					time.Sleep(waitTime)
					continue
				} else {
					log.Printf("发送请求失败: %v", err)
					break
				}
			} else {
				// 请求成功，跳出重试循环
				break
			}
		}

		// 请求间隔，避免频率过快
		if i < len(rows)-1 { // 不是最后一个请求
			fmt.Printf("等待 %v 后发送下一个请求...\n", requestInterval)
			time.Sleep(requestInterval)
		}
	}

	fmt.Println("所有测试请求已发送完成")
}

// 读取CSV文件
func readCSV(filename string) ([]CSVRow, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV文件至少需要包含表头和一行数据")
	}

	var rows []CSVRow
	for i := 1; i < len(records); i++ { // 跳过表头
		record := records[i]
		if len(record) < 4 {
			continue
		}

		appID, err := strconv.Atoi(record[0])
		if err != nil {
			log.Printf("解析app_id失败: %v", err)
			continue
		}

		// 解析scenes字段，假设用逗号分隔
		scenes := strings.Split(record[2], ";")
		for j := range scenes {
			scenes[j] = strings.TrimSpace(scenes[j])
		}

		rows = append(rows, CSVRow{
			AppID:    appID,
			RoleDesc: strings.TrimSpace(record[1]),
			Scenes:   scenes,
			Style:    strings.TrimSpace(record[3]),
		})
	}

	return rows, nil
}

// 发送测试请求
func sendTestRequest(row CSVRow) error {
	req := TestRequest{
		AppID:    row.AppID,
		RoleDesc: row.RoleDesc,
		Scenes:   row.Scenes,
		Style:    row.Style,
		Callback: callbackURL,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response TestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	if response.Code != 0 {
		return fmt.Errorf("API返回错误: %s", response.Msg)
	}

	// 创建task_id文件夹并保存描述信息
	taskFolder := filepath.Join(testFolder, response.TaskID)
	if err := os.MkdirAll(taskFolder, 0755); err != nil {
		return err
	}

	// 保存请求信息到desc.txt
	descContent := fmt.Sprintf("Role Description: %s\nScenes: %s\n", row.RoleDesc, strings.Join(row.Scenes, ", "))
	descFile := filepath.Join(taskFolder, "desc.txt")
	if err := os.WriteFile(descFile, []byte(descContent), 0644); err != nil {
		return err
	}

	fmt.Printf("任务创建成功: %s\n", response.TaskID)
	return nil
}

// 启动Web服务器
func startWebServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// 处理回调
	r.POST("/callback", handleCallback)

	// 图片文件服务 - 直接映射测试文件夹
	r.StaticFS("/files", http.Dir("./"))

	// Web界面路由
	r.GET("/", handleIndex)
	r.GET("/test/:testName", handleTestDetail)
	r.GET("/task/:testName/:taskID", handleTaskDetail)

	fmt.Println("Web服务器启动在 http://localhost:9983")
	log.Fatal(r.Run(":9983"))
}

// 处理回调
func handleCallback(c *gin.Context) {
	var callback CallbackRequest
	if err := c.ShouldBindJSON(&callback); err != nil {
		c.JSON(400, gin.H{"error": "解析回调数据失败"})
		return
	}

	fmt.Printf("收到回调: TaskID=%s, Success=%v\n", callback.TaskID, callback.Success)

	if callback.Msg != "success" { // 不要动
		log.Printf("任务失败: %s, 错误信息: %s", callback.TaskID, callback.Msg)
		c.JSON(200, gin.H{"status": "ok"})
		return
	}

	// 找到对应的task文件夹
	taskFolder := ""
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), "test-") {
			taskPath := filepath.Join(path, callback.TaskID)
			if _, err := os.Stat(taskPath); err == nil {
				taskFolder = taskPath
				return filepath.SkipDir
			}
		}
		return nil
	})

	if err != nil || taskFolder == "" {
		log.Printf("找不到TaskID对应的文件夹: %s", callback.TaskID)
		c.JSON(200, gin.H{"status": "ok"})
		return
	}

	// 读取场景信息
	descFile := filepath.Join(taskFolder, "desc.txt")
	descContent, err := os.ReadFile(descFile)
	if err != nil {
		log.Printf("读取描述文件失败: %v", err)
		c.JSON(200, gin.H{"status": "ok"})
		return
	}

	// 解析场景列表
	scenes := extractScenesFromDesc(string(descContent))

	// 下载图片
	for _, img := range callback.Data.Imgs {
		sceneName := ""
		if img.Index < len(scenes) {
			sceneName = scenes[img.Index]
		} else {
			sceneName = fmt.Sprintf("原图_%d", img.Index)
		}

		if err := downloadImage(img.URL, taskFolder, sceneName); err != nil {
			log.Printf("下载图片失败: %v", err)
		}
	}

	c.JSON(200, gin.H{"status": "ok"})
}

// 从描述文件中提取场景列表
func extractScenesFromDesc(content string) []string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Scenes:") {
			scenesStr := strings.TrimPrefix(line, "Scenes:")
			scenesStr = strings.TrimSpace(scenesStr)
			return strings.Split(scenesStr, ", ")
		}
	}
	return []string{}
}

// 下载图片
func downloadImage(url, folder, sceneName string) error {
	// 如果URL不包含协议前缀，添加默认前缀
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://res.theact.ai/" + url
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 从URL中获取文件扩展名，如果没有则默认使用.jpg
	ext := ".jpg"
	if strings.Contains(url, ".") {
		parts := strings.Split(url, ".")
		if len(parts) > 1 {
			lastPart := parts[len(parts)-1]
			// 移除可能的查询参数
			if idx := strings.Index(lastPart, "?"); idx != -1 {
				lastPart = lastPart[:idx]
			}
			if lastPart != "" {
				ext = "." + lastPart
			}
		}
	}

	// 创建文件
	filename := fmt.Sprintf("%s%s", sceneName, ext)
	filepath := filepath.Join(folder, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 复制数据
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("图片下载成功: %s\n", filepath)
	return nil
}

// 处理首页
func handleIndex(c *gin.Context) {
	testFolders := getTestFolders()
	c.HTML(200, "index.html", gin.H{
		"TestFolders": testFolders,
	})
}

// 处理测试详情页
func handleTestDetail(c *gin.Context) {
	testName := c.Param("testName")
	tasks := getTaskFolders(testName)
	c.HTML(200, "test_detail.html", gin.H{
		"TestName": testName,
		"Tasks":    tasks,
	})
}

// 处理任务详情页
func handleTaskDetail(c *gin.Context) {
	testName := c.Param("testName")
	taskID := c.Param("taskID")

	taskFolder := filepath.Join(testName, taskID)
	images := getImages(taskFolder)
	desc := getDescription(taskFolder)

	c.HTML(200, "task_detail.html", gin.H{
		"TestName":    testName,
		"TaskID":      taskID,
		"Images":      images,
		"Description": desc,
	})
}

// 获取测试文件夹列表
func getTestFolders() []string {
	var folders []string
	entries, err := os.ReadDir(".")
	if err != nil {
		return folders
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "test-") {
			folders = append(folders, entry.Name())
		}
	}
	return folders
}

// 获取任务文件夹列表
func getTaskFolders(testName string) []string {
	var tasks []string
	entries, err := os.ReadDir(testName)
	if err != nil {
		return tasks
	}

	for _, entry := range entries {
		if entry.IsDir() {
			tasks = append(tasks, entry.Name())
		}
	}
	return tasks
}

// 获取图片列表
func getImages(taskFolder string) []string {
	var images []string
	entries, err := os.ReadDir(taskFolder)
	if err != nil {
		return images
	}

	// 支持的图片格式
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}

	for _, entry := range entries {
		if !entry.IsDir() {
			filename := strings.ToLower(entry.Name())
			for _, ext := range imageExts {
				if strings.HasSuffix(filename, ext) {
					images = append(images, entry.Name())
					break
				}
			}
		}
	}
	return images
}

// 获取描述信息
func getDescription(taskFolder string) string {
	descFile := filepath.Join(taskFolder, "desc.txt")
	content, err := os.ReadFile(descFile)
	if err != nil {
		return "无描述信息"
	}
	return string(content)
}
