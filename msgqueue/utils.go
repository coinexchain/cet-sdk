package msgqueue

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetFilePathAndFileIndexFromDir(dir string, maxFileSize int) (filePath string, fileIndex int, err error) {
	if _, fileIndex, err = getFilePathAndIndex(dir, maxFileSize); err != nil {
		return
	}

	fileSize := GetFileSize(GetFileName(dir, fileIndex))
	if fileSize < maxFileSize {
		return GetFileName(dir, fileIndex), fileIndex, nil
	}
	return GetFileName(dir, fileIndex+1), fileIndex + 1, nil
}

func getFilePathAndIndex(dir string, height int) (filePath string, fileIndex int, err error) {
	fileNames, err := getAllFilesFromDir(dir)
	if err != nil {
		return "", -1, err
	}
	if len(fileNames) == 0 {
		return GetFileName(dir, 0), 0, nil
	}
	fileIndex = getMaxIndexFromFiles(fileNames)
	return GetFileName(dir, fileIndex), fileIndex, nil
}

func getAllFilesFromDir(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			return nil, err
		}
		return nil, nil
	}

	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			if strings.HasPrefix(name, filePrefix) {
				fileNames = append(fileNames, name)
			}
		}
	}
	return fileNames, nil
}

func getMaxIndexFromFiles(fileNames []string) int {
	fileIndex := 0
	for _, fileName := range fileNames {
		vals := strings.Split(fileName, "-")
		if len(vals) == 2 {
			if index, err := strconv.Atoi(vals[1]); err == nil {
				if index > fileIndex {
					fileIndex = index
				}
			}
		}
	}
	return fileIndex
}

func getMinIndexFromFiles(fileNames []string) int {
	fileIndex := math.MaxInt64
	for _, fileName := range fileNames {
		vals := strings.Split(fileName, "-")
		if len(vals) == 2 {
			if index, err := strconv.Atoi(vals[1]); err == nil {
				if index < fileIndex {
					fileIndex = index
				}
			}
		}
	}
	return fileIndex
}

func GetFileName(dir string, fileIndex int) string {
	return dir + "/" + filePrefix + strconv.Itoa(fileIndex)
}

func GetFileSize(filePath string) int {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) || info.IsDir() {
		return -1
	}

	return int(info.Size())
}

func GetFileLeastHeightInDir(dir string) (string, int64, error) {
	files, err := getAllFilesFromDir(dir)
	if err != nil {
		return "", -1, err
	}
	index := getMinIndexFromFiles(files)
	in, err := openFile(GetFileName(dir, index))
	if err != nil {
		return "", -1, err
	}
	defer in.Close()

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil {
		return "", -1, err
	}
	return GetFileName(dir, index), getHeight(line), nil
}

func getHeight(data string) int64 {
	if len(data) == 0 {
		return -1
	}
	vals := strings.Split(data, "#")
	if vals[0] != "height_info" {
		panic("The first line of data in the file should be [height_info] msg")
	}
	var info NewHeightInfo
	err := json.Unmarshal([]byte(vals[1]), &info)
	if err != nil {
		panic(fmt.Sprintf("json unmarshal error : %s\n", err.Error()))
	}
	return info.Height
}

func openFile(filePath string) (*os.File, error) {
	if s, err := os.Stat(filePath); os.IsNotExist(err) {
		return os.Create(filePath)
	} else if s.IsDir() {
		return nil, fmt.Errorf("Need to give the file path ")
	} else {
		return os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0666)
	}
}

func FillMsgs(ctx sdk.Context, key string, msg interface{}) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeMsgQueue, sdk.NewAttribute(key, string(bytes))))
}
