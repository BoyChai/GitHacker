package recovery

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var IndexMagic = [4]byte{0x44, 0x49, 0x52, 0x43}

// indexStr 结构体定义
type indexStr struct {
	Sha    string
	Length int
	Name   string
}

// 移除不可见字符
func removeControlCharacters(s string) string {
	var result strings.Builder
	for _, char := range s {
		if char >= 32 {
			result.WriteRune(char)
		}
	}
	return result.String()
}

func saveFile(filePath string, data string) {
	oldPath := filePath
	filePath = filepath.Clean(OutputDir + "/" + filePath)
	filePath = removeControlCharacters(filePath)
	dir := filepath.Dir(filePath) // 获取文件的目录路径

	// 创建必要的目录
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("[x]  %s, err: %v\n", filePath, err)
		return
	}

	// 创建文件（如果文件不存在，创建新文件；如果文件已存在，覆盖）
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("[x]  %s, err: %v\n", filePath, err)
		return
	}
	defer file.Close()

	// 将数据写入文件
	_, err = file.WriteString(data)
	if err != nil {
		fmt.Printf("[x]  %s, err: %v\n", filePath, err)
		return
	}

	// 成功输出
	fmt.Printf("[√] %s\n", oldPath)
}

// Indexs 用于存储索引信息
var Indexs []indexStr

// nextMultipleOf8 返回大于等于n且为8的倍数的数
// 用于index文件条目不可控位数的计算
func nextMultipleOf8(n int) int {
	n = n + 62
	if n%8 == 0 {
		return n - 62 + 8
	}
	return n + (8 - n%8) - 62
}

// readIndex 读取索引信息
func readIndex(index io.Reader) {
	var magic [4]byte
	if err := binary.Read(index, binary.BigEndian, &magic); err != nil {
		fmt.Println("[x] Error reading magic number:", err)
		return
	}
	if magic != IndexMagic {
		fmt.Println("文件头错误，请检查URL是否正确")
		os.Exit(1)
	}
	var version uint32
	if err := binary.Read(index, binary.BigEndian, &version); err != nil {
		fmt.Println("[x] Error reading version:", err)
		return
	}
	fmt.Printf("Git文件格式版本: %d\n", version)
	if version != 2 {
		fmt.Println("git版本太旧或者太新了,请联系开发者")
		os.Exit(1)
	}
	var entryCount uint32
	if err := binary.Read(index, binary.BigEndian, &entryCount); err != nil {
		fmt.Println("[x] Error reading entry count:", err)
		return
	}
	fmt.Printf("文件总数: %d\n", entryCount)

	// 逐条读取索引条目
	for i := uint32(0); i < entryCount; i++ {
		var entry [62]byte
		if err := binary.Read(index, binary.BigEndian, &entry); err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println("[x] Error reading entry:", err)
			return
		}

		// 提取文件信息
		sha1 := entry[40:60]
		fileLength := entry[61:62]
		indexLength := nextMultipleOf8(int(fileLength[0]))

		fileName := make([]byte, indexLength)
		if err := binary.Read(index, binary.BigEndian, &fileName); err != nil {
			if err.Error() == "EOF" {
				fmt.Println("总文件数:", i)
				break
			}
			fmt.Println("[x] Error reading entry:", err)
			return
		}

		Indexs = append(Indexs, indexStr{
			Sha:    fmt.Sprintf("%x", sha1),
			Length: int(fileLength[0]),
			Name:   string(fileName),
		})
		// fmt.Printf("Entry %d: Sha-1 %x Length %d File: %s\n", i, sha1, fileLength[0], fileName)
	}
}

// fetchURL 统一处理URL请求和错误
func fetchURL(url string) (*http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// UrlRecovery 从URL恢复数据
func UrlRecovery(url string) {

	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	// 读取索引
	r, err := fetchURL(url + "index")
	if err != nil {
		fmt.Println("[x] Error fetching index:", err)
		return
	}
	defer r.Body.Close()
	readIndex(r.Body)

	// 创建任务通道
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU()*2)

	// 逐个下载并解压每个对象
	for _, i := range Indexs {
		objectURL := fmt.Sprintf("%sobjects/%s/%s", url, i.Sha[:2], i.Sha[2:])

		wg.Add(1)         // 增加WaitGroup计数
		sem <- struct{}{} // 获取一个信号量，限制并发

		go func(i indexStr) {
			defer wg.Done()
			defer func() { <-sem }()

			r, err := fetchURL(objectURL)
			if err != nil {
				fmt.Println("[x] Error fetching object:", err)
				return
			}
			defer r.Body.Close()

			// zlib解压
			reader, err := zlib.NewReader(r.Body)
			if err != nil {
				fmt.Println("[x]", i.Name, err)
				return
			}
			defer reader.Close()

			data, err := io.ReadAll(reader)
			zeroIndex := bytes.IndexByte(data, 0x00)
			if zeroIndex != -1 {
				// 丢弃 0x00 前的部分，只保留后面的内容
				data = data[zeroIndex+1:]
			}
			if err != nil {
				fmt.Println("[x]", err)
				return
			}
			saveFile(i.Name, string(data))
		}(i)
	}

	wg.Wait()
	fmt.Println("Sir My task has been completed！")
}
