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

func main() {
	fmt.Println(os.Args)
	var targetDirs []string = []string{}
	if len(os.Args) < 2 {
		targetDirs = append(targetDirs, ".")
	} else {
		targetDirs = append(targetDirs, os.Args[1:]...)
	}

	root := targetDirs[0] // 保存位置

	for _, dir := range targetDirs {
		subDirs, err := getAllSubDir(dir)
		if err != nil {
			panic(err)
		}
		for _, sd := range subDirs {
			subDir := strings.ReplaceAll(sd, "\n", "")
			entryPath := path.Join(subDir, "entry.json")
			fmt.Println(entryPath)
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
				if err := merge(subVPath, path.Join(savePath, formatDirectoryName(entry.Pagedata.Part))); err != nil {
					switch err.(type) {
					case *exec.ExitError:
						// 存在的跳过
						continue
					default:
						panic(err)
					}
				}
			}
		}
	}
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
	tmp := "'" + strings.TrimSpace(name) + "'"
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
			fmt.Println(tmp)
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
		if err == os.ErrNotExist {
			return nil, err
		}
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
