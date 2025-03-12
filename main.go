package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

type PageData struct {
	Page int    `json:"page"`
	Part string `json:"part"`
}
type Entry struct {
	Pagedata  PageData `json:"page_data"`
	OwnerName string   `json:"owner_name"`
	Title     string   `json:"title"`
	Cover     string   `json:"cover"`
	TypeTag   string   `json:"type_tag"`
}

var version string = "0.1.2"

func main() {

	if len(os.Args) > 1 && (strings.EqualFold(os.Args[1], "--help") || strings.EqualFold(os.Args[1], "-h")) {
		printHelp()
		os.Exit(0)
	}

	var targetDirs []string = []string{}
	if len(os.Args) < 2 {
		// 获取当前目录下所有目录作为参数
		subCurDirs, err := getAllSubDir(".")
		if err != nil {
			fmt.Printf("error: %v", err)
			os.Exit(1)
		}
		targetDirs = append(targetDirs, subCurDirs...)
	} else {
		targetDirs = append(targetDirs, os.Args[1:]...)
	}

	root := targetDirs[0] // 保存位置
	for _, dir0 := range targetDirs {
		dir := strings.ReplaceAll((dir0), "\n", "")
		subDirs, err := getAllSubDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(dir, "无效的路径")
				os.Exit(0)
			} else {
				panic(err)
			}
		}
		for k, dir1 := range subDirs {
			subDir := strings.ReplaceAll(dir1, "\n", "")
			entryPath := path.Join(subDir, "entry.json")
			if !isFileExist(entryPath) {
				//fmt.Printf("%v 不存在entry.json,跳过\n", dir1)
				continue
			}
			entry, err := getJsonFileContent[Entry](entryPath)
			savePath := path.Join(path.Dir(root), formatDirectoryName(entry.Title))
			// 判断是否存在已生成目录
			if isFileExist(savePath) {
				fmt.Println(entry.Title, "已经存在同名目录，跳过。")
				continue
			}
			if err != nil {
				panic(err)
			}
			if entry.TypeTag != "" {
				subVPath := path.Join(subDir, entry.TypeTag)
				// 确保输出目录存在
				if err := makesureDirExist(savePath); err != nil {
					panic(err)
				}
				// 执行合成输出
				partFileName := fmt.Sprintf("P%v-%v", entry.Pagedata.Page, entry.Pagedata.Part)
				if err := merge(subVPath, path.Join(savePath, formatDirectoryName(partFileName))); err != nil {
					switch err.(type) {
					case *exec.ExitError:
						// 存在的跳过
						continue
					default:
						panic(err)
					}
				}
				fmt.Printf("\r(%v/%v): %v", k+1, len(subDirs), entry.Title)
			}
		}
		fmt.Println("")
	}
}

func printHelp() {
	fmt.Printf("bili-merge v%v \n", version)
	fmt.Println("用法: bili-merge [参数...]")
	fmt.Println("参数:")
	fmt.Println("  参数...             传递目录路径，可以传递多个")
	fmt.Println("  -h,--help          显示此帮助信息并退出")
}

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return true
}

// 防止目录名带有特殊符号
func formatDirectoryName(name string) string {
	tmp := "'" + strings.ReplaceAll(name, " ", "") + "'"
	//tmp = strings.Replace(tmp, "\\n", "", -1)
	return tmp
}

// 使用ffmpeg合并audio.m4s和video.m4s
func merge(dir string, target string) error {
	file1 := path.Join(dir, "video.m4s")
	file2 := path.Join(dir, "audio.m4s")
	cmdStr := fmt.Sprintf("ffmpeg -i %v -i %v -codec copy %v.mp4", file1, file2, target)
	// fmt.Println(cmdStr)
	cmd := exec.Command("sh", "-c", cmdStr)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func makesureDirExist(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			tmp := fmt.Sprintf("mkdir -p %v", dir)
			cmd := exec.Command("sh", "-c", tmp)
			err := cmd.Run()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// func renameDir(oldDir string, newName string) error {
// 	if oldDir == "" {
// 		return fmt.Errorf("oldDir not found: %v", oldDir)
// 	}

// 	cmd := exec.Command("sh", "-c", fmt.Sprintf("mv %v %v", oldDir, path.Join(path.Dir(oldDir), newName)))
// 	if err := cmd.Run(); err != nil {
// 		return err
// 	}
// 	return nil
// }

func getJsonFileContent[T any](filepath string) (*T, error) {
	file, err := os.OpenFile(filepath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bs, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(bs, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func getAllSubDir(dir string) ([]string, error) {
	// 判断文件夹是否存在
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, nil
	}
	cmd := exec.Command("sh", "-c", fmt.Sprintf("find %v -path '*/.*' -prune -o -type d -mindepth 1 -maxdepth 1 -print0 | xargs -0 echo", dir))
	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	outStr, errStr := stdout.String(), stderr.String()
	if errStr != "" {
		return nil, fmt.Errorf("命令执行出错，error：%v", errStr)
	}
	return strings.Split(outStr, " "), nil
}
