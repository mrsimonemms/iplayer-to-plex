package convert

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type FilePath struct {
	Dir  string
	Info os.FileInfo
	Path string
}

type Programme struct {
	Programme struct {
		Position     int    `json:"position"`
		Title        string `json:"title"`
		DisplayTitle struct {
			Title string `json:"title"`
		} `json:"display_title"`
		Parent struct {
			Programme struct {
				Position int    `json:"position"`
				Title    string `json:"title"`
			} `json:"programme"`
		} `json:"parent"`
	} `json:"programme"`
}

func (obj *Programme) NewName(ext string) string {
	episodeTitle := removeNoAlnum(obj.Programme.Title)
	showTitle := removeNoAlnum(obj.Programme.DisplayTitle.Title)
	episodeNumber := obj.Programme.Position
	seriesNumber := obj.Programme.Parent.Programme.Position

	var name string
	if episodeNumber == 0 && seriesNumber == 0 {
		/* Treat as a single programme */
		name = episodeTitle
	} else {
		name = fmt.Sprintf(
			"%s - s%se%s - %s",
			showTitle,
			leftPad(seriesNumber),
			leftPad(episodeNumber),
			episodeTitle,
		)
	}

	name += ext

	return name
}

func (obj *Programme) DirName() string {
	showTitle := removeNoAlnum(obj.Programme.DisplayTitle.Title)
	episodeNumber := obj.Programme.Position
	seriesNumber := obj.Programme.Parent.Programme.Position

	var name string
	if episodeNumber != 0 && seriesNumber != 0 {
		name = filepath.Join(showTitle, fmt.Sprintf("Series %d", seriesNumber))
	}

	return name
}

func Convert(dir string, setDir bool) (int, error) {
	files, err := getFiles(dir)

	count := 0

	if err != nil {
		return count, err
	}

	for _, file := range files {
		pid, err := getPid(file.Info.Name())

		if err != nil {
			fmt.Println(err)
			continue
		}

		if pid == nil {
			fmt.Println("File aleady converted: " + file.Path)
			continue
		}

		ext := filepath.Ext(file.Info.Name())

		programme, err := getNewName(*pid)

		if err != nil {
			fmt.Println(err)
			continue
		}

		name := programme.NewName(ext)
		dir := filepath.Dir(file.Path)

		if setDir {
			dir = filepath.Join(dir, "../", programme.DirName())
		}

		/* Ensure that the directories exist */
		os.MkdirAll(dir, 0755)

		newFilePath := filepath.Join(dir, name)

		fmt.Printf("Converting \"%s\" to \"%s\"\n", file.Info.Name(), name)
		err = os.Rename(file.Path, newFilePath)

		if err != nil {
			fmt.Println(err)
			continue
		}

		empty, err := isEmpty(file.Dir)

		if err != nil {
			fmt.Println(err)
			continue
		}

		if empty {
			/* Remove directory if it's empty */
			fmt.Printf("Removing directory %s\n", file.Dir)
			err := os.Remove(file.Dir)

			if err != nil {
				fmt.Println(err)
			}
		}

		count += 1
	}

	return count, nil
}

func getFiles(dir string) ([]FilePath, error) {
	var files []FilePath

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, FilePath{
				Dir:  filepath.Dir(path),
				Info: info,
				Path: path,
			})
		}

		return nil
	})

	return files, err
}

func getPid(fileName string) (*string, error) {
	re := regexp.MustCompile("(\\w+)\\s(original|editorial|podcast|iplayer)\\..*$")
	match := re.FindStringSubmatch(fileName)

	if len(match) == 0 {
		err := errors.New("Cannot find PID in " + fileName)

		return nil, err
	}

	return &match[1], nil
}

func getNewName(pid string) (*Programme, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url := "https://www.bbc.co.uk/programmes/" + pid + ".json"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("Unknown PID: " + pid)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	data := Programme{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func leftPad(i int) string {
	var x string

	if i < 10 {
		x += "0"
	}

	x += strconv.Itoa(i)

	return x
}

func removeNoAlnum(str string) string {
	re := regexp.MustCompile("\\W")
	return re.ReplaceAllStringFunc(str, func(s string) string {
		if s != " " {
			return ""
		}

		return s
	})
}
