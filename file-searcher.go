// File Searcher
// V1.0

package main

import (
    "flag"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "sync"
    "time"
)

var (
    subname string // 文件后缀名
    keyword string // 关键词
    date    int    // 修改日期范围（天）
    dir     string // 扫描目录
    help    bool   // 帮助信息标志
)

const (
    colorBlue  = "\033[34m"
    colorReset = "\033[0m"
)

type FileInfo struct {
    Path         string    // 文件路径
    Name         string    // 文件名
    ModifiedTime time.Time // 修改时间
    IsDir        bool      // 是否为目录
    Depth        int       // 目录深度
}

// 工作池结构体，用于控制并发数量
type WorkerPool struct {
    wg      sync.WaitGroup
    jobChan chan FileInfo
    results chan FileInfo
    maxJobs int
}

// 创建新的工作池
func NewWorkerPool(maxJobs int) *WorkerPool {
    return &WorkerPool{
        jobChan: make(chan FileInfo, maxJobs),
        results: make(chan FileInfo, maxJobs),
        maxJobs: maxJobs,
    }
}

// 启动工作池
func (wp *WorkerPool) Start() {
    for i := 0; i < wp.maxJobs; i++ {
        wp.wg.Add(1)
        go func() {
            defer wp.wg.Done()
            for job := range wp.jobChan {
                if !job.IsDir && isMatchingFile(job) {
                    wp.results <- job
                }
            }
        }()
    }
}

func (wp *WorkerPool) AddJob(job FileInfo) {
    wp.jobChan <- job
}

func (wp *WorkerPool) Close() {
    close(wp.jobChan)
    wp.wg.Wait()
    close(wp.results)
}

// 判断文件是否符合过滤条件
func isMatchingFile(file FileInfo) bool {
    if subname != "" {
        fileExt := strings.TrimPrefix(filepath.Ext(file.Name), ".")
        if !strings.EqualFold(fileExt, subname) {
            return false
        }
    }

    if date > 0 {
        dateLimit := time.Now().AddDate(0, 0, -date)
        if file.ModifiedTime.Before(dateLimit) {
            return false
        }
    }

    return true
}

func getRelativeTimeDesc(t time.Time) string {
    now := time.Now()
    duration := now.Sub(t)

    switch {
    case duration < time.Minute:
        return "刚刚"
    case duration < time.Hour:
        return fmt.Sprintf("%d 分钟前", int(duration.Minutes()))
    case duration < 24*time.Hour:
        return fmt.Sprintf("%d 小时前", int(duration.Hours()))
    case duration < 30*24*time.Hour:
        return fmt.Sprintf("%d 天前", int(duration.Hours()/24))
    case duration < 365*24*time.Hour:
        return fmt.Sprintf("%d 个月前", int(duration.Hours()/(24*30)))
    default:
        return fmt.Sprintf("%d 年前", int(duration.Hours()/(24*365)))
    }
}

func formatFileName(name string) string {
    if keyword != "" && strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) {
        return colorBlue + name + colorReset
    }
    return name
}

func printFileList(files []FileInfo) {
    for _, file := range files {
        relativeTime := getRelativeTimeDesc(file.ModifiedTime)
        fmt.Printf("· %s - %s\n", formatFileName(file.Name), relativeTime)
    }
}

func init() {
    flag.StringVar(&subname, "s", "", "指定后缀名（例如：mp4、txt、md）")
    flag.StringVar(&subname, "subname", "", "指定后缀名（例如：mp4、txt、md）")

    flag.StringVar(&keyword, "k", "", "指定特别关键词")
    flag.StringVar(&keyword, "keyword", "", "指定特别关键词")

    flag.IntVar(&date, "d", 0, "指定修改日期范围（天）")
    flag.IntVar(&date, "date", 0, "指定修改日期范围（天）")

    flag.StringVar(&dir, "i", "", "指定要搜索的目录")
    flag.StringVar(&dir, "input", "", "指定要搜索的目录")

    flag.BoolVar(&help, "h", false, "显示帮助信息")
    flag.BoolVar(&help, "help", false, "显示帮助信息")

    flag.Usage = func() {
        fmt.Println("Usage: filesearcher [options] [directory]")
        fmt.Println("Options:")
        fmt.Println("  -s, --subname string   指定后缀名（例如：mp4、txt、md）")
        fmt.Println("  -k, --keyword string   指定特别关键词")
        fmt.Println("  -d, --date int         指定修改日期范围（天）")
        fmt.Println("  -i, --input string     指定要搜索的目录")
        fmt.Println("  -h, --help             显示帮助信息")
    }
}

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Copyright (C) 2025 Yule")
        fmt.Println("GitHub: https://github.com/YuleBest")
        fmt.Println("--- File Searcher V1.0 ---")
        return
    }

    flag.Parse()

    if help {
        flag.Usage()
        return
    }

    if dir == "" {
        args := flag.Args()
        if len(args) > 0 {
            dir = args[0]
        } else {
            var err error
            dir, err = os.Getwd()
            if err != nil {
                fmt.Println("获取当前目录失败:", err)
                os.Exit(1)
            }
        }
    }

    workerPool := NewWorkerPool(10)
    workerPool.Start()

    var results []FileInfo
    var resultMutex sync.Mutex
    var wg sync.WaitGroup

    // 启动结果收集协程
    wg.Add(1)
    go func() {
        defer wg.Done()
        for result := range workerPool.results {
            resultMutex.Lock()
            results = append(results, result)
            resultMutex.Unlock()
        }
    }()

    // 遍历目录
    err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // 获取文件信息
        info, err := d.Info()
        if err != nil {
            return nil // 跳过无法获取信息的文件
        }

        relPath, err := filepath.Rel(dir, path)
        if err != nil {
            relPath = path
        }
        if relPath == "." {
            return nil // 跳过根目录
        }

        depth := len(strings.Split(relPath, string(os.PathSeparator))) - 1
        if d.IsDir() {
            depth--
        }

        // 创建文件信息
        fileInfo := FileInfo{
            Path:         path,
            Name:         d.Name(),
            ModifiedTime: info.ModTime(),
            IsDir:        d.IsDir(),
            Depth:        depth,
        }

        if !d.IsDir() {
            workerPool.AddJob(fileInfo)
        }

        return nil
    })

    // 关闭工作池并等待任务完成
    workerPool.Close()
    wg.Wait()

    if err != nil {
        fmt.Println("扫描目录失败:", err)
        os.Exit(1)
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].ModifiedTime.After(results[j].ModifiedTime)
    })

    if len(results) == 0 {
        fmt.Println("未找到匹配的文件")
        return
    }

    fmt.Printf("在 %s 目录下找到 %d 个匹配的文件:\n\n", dir, len(results))
    printFileList(results)
}