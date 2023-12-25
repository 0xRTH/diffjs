package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func getHash(body []byte) []byte {
	h := sha256.New()
	h.Write(body)
	return h.Sum(nil)
}

func getAllFilesIn(path string) []string {
	list := []string{}
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err.Error())
		}
		list = append(list, info.Name())
		return nil
	})
	return list
}

func downloadAndHash(url string) {
	fmt.Printf("Getting: %s\n", url)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal("Could not get", url, "\n", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Could not get ", url, "\n", err)
	}

	file, err := os.Create(fmt.Sprintf("./temp_Downloads/%x_%s", getHash(body), strings.ReplaceAll(strings.TrimPrefix(url, "https://"), "/", "_")))
	if err != nil {
		log.Fatal("could not create file :", err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	_, err = w.WriteString(string(body))
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()
}

func checkForNewHashes(oldPath string, newPath string) ([]string, [][]string) {
	oldHashes := make(map[string]string)
	newHashes := make(map[string]string)
	newJSFiles := []string{}
	diffJSFiles := [][]string{}

	// get a list of all new hashes - Same
	newFiles := getAllFilesIn(newPath)
	for _, file := range newFiles[1:] {
		newHashes[strings.SplitN(file, "_", 2)[0]] = file // strings.SplitN(file, "_", 2)[1]
	}
	// get a list of all old hashes - map to still get the name of the file ?
	_, err := os.Stat(oldPath)
	if os.IsNotExist(err) {
		fmt.Println("No old folder to check against provided")
		fmt.Println(newFiles)
		return newFiles[1:], [][]string{}
	}
	oldFiles := getAllFilesIn(oldPath)
	for _, file := range oldFiles[1:] {
		oldHashes[strings.SplitN(file, "_", 2)[0]] = file // strings.SplitN(file, "_", 2)[1]
	}

	//get either file as new or diff
	for hash, urlNew := range newHashes {
		_, ok := oldHashes[hash]
		if !ok {
			isFileNew := true
			for _, urlOld := range oldHashes {
				if strings.SplitN(urlNew, "_", 2)[1] == strings.SplitN(urlOld, "_", 2)[1] {
					diffJSFiles = append(diffJSFiles, []string{urlNew, urlOld})
					isFileNew = false
				}
			}
			if isFileNew {
				newJSFiles = append(newJSFiles, urlNew)
			}
		}
	}
	return newJSFiles, diffJSFiles
}

func getDiff(path1 string, path2 string) []byte {
	diff, _ := exec.Command("diff", "-c", path1, path2).Output()
	// if err != nil {
	// 	log.Fatal("Error while executing os command: ", err)
	// }
	return diff
}

func saveLogAndReturnShortLog(dir string, newJsFiles []string, allDiffs map[[2]string][]byte) string {
	shortLog := ""

	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
	dt := time.Now()
	file, err := os.Create(fmt.Sprintf("%s/log_%s", dir, dt.Format(time.RFC3339)))
	defer file.Close()

	w := bufio.NewWriter(file)
	_, err = w.WriteString("New files : \n\n")
	shortLog += "New files : \n\n"
	for _, new := range newJsFiles {
		_, err = w.WriteString(fmt.Sprintf("%s\n", nameToUrl(new)))
		shortLog += fmt.Sprintf("%s\n", nameToUrl(new))
	}
	_, err = w.WriteString("\n\nAll diffs :")
	shortLog += fmt.Sprintf("\n\n%d changes :\n", len(allDiffs))
	for files, diff := range allDiffs {
		_, err = w.WriteString(fmt.Sprintf("\n\n%s & %s:\n\n", files[0], files[1]))
		_, err = w.WriteString(string(diff))
		shortLog += fmt.Sprintf("\n%s -> %s:\n\n", files[0], files[1])
	}
	w.Flush()
	return shortLog
}

func notify(data string, notifyId string) {
	fmt.Println("Notifying...")
	if notifyId != "" {
		_ = exec.Command("sh", "-c", fmt.Sprintf("echo \"%s\" | notify -bulk -id %s", data, notifyId)).Run()
	} else {
		_ = exec.Command("sh", "-c", fmt.Sprintf("echo \"%s\" | notify -bulk", data)).Run()
	}
}

func nameToUrl(name string) string {
	url := "https://"
	url += strings.SplitN(name, "_", 2)[1]
	url = strings.ReplaceAll(url, "_", "/")
	return url
}

func main() {
	archiveDir := flag.String("archive", "./Archives", "Set folder for archives (default:'./Archives')")
	oldDir := flag.String("dir", "./Downloads", "Set folder where previous files are (default:'./Downloads')")
	newDir := "./temp_Downloads"
	logsDir := flag.String("log", "./Logs", "Set folder for storing logs (default:'./Logs')")
	isNotify := flag.Bool("notify", false, "Activate notification via notify")
	nId := flag.String("nId", "", "Notify id to send the notification to (let empty to notify to default)")
	var urlsList []string
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		urlsList = append(urlsList, sc.Text())
	}

	flag.Parse()

	_, err := os.Stat(newDir)
	if os.IsNotExist(err) {
		os.Mkdir(newDir, 0755)
	}

	for _, url := range urlsList {
		downloadAndHash(url)
	}

	allDiffs := make(map[[2]string][]byte)

	newJsFiles, diffJsFiles := checkForNewHashes(*oldDir, newDir)
	for _, files := range diffJsFiles {
		allDiffs[[2]string{files[0], files[1]}] = getDiff(fmt.Sprintf("%s/%s", newDir, files[0]), fmt.Sprintf("%s/%s", *oldDir, files[1]))
	}

	fmt.Printf("Downloaded %d files in %s\n", len(urlsList), *oldDir)
	fmt.Printf("%d new files found\n", len(newJsFiles))
	fmt.Printf("%d files with modifications found\n", len(diffJsFiles))

	_, err = os.Stat(*archiveDir)
	if os.IsNotExist(err) {
		os.Mkdir(*archiveDir, 0755)
	}

	err = exec.Command("sh", "-c", fmt.Sprintf("mv -n %s/* %s", *oldDir, *archiveDir)).Run()
	err = exec.Command("sh", "-c", fmt.Sprintf("rm -rf %s; mv %s/ %s", *oldDir, newDir, *oldDir)).Run()
	if len(newJsFiles) != 0 || len(allDiffs) != 0 {
		shortLog := saveLogAndReturnShortLog(*logsDir, newJsFiles, allDiffs)
		if *isNotify {
			notify(shortLog, *nId)
			// fmt.Println(shortLog, nId)
		}
	}

}

// a faire :
// Rajouter les arguments : pour les output (-s, et -v)
