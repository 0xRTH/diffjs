package main

import (
	"bufio"
	"crypto/sha256"
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

	// get a list of all old hashes - map to still get the name of the file ?
	_, err := os.Stat(oldPath)
	if os.IsNotExist(err) {
		fmt.Println("No old folder to check against provided")
		return []string{}, [][]string{}
	}
	oldFiles := getAllFilesIn(oldPath)
	for _, file := range oldFiles[1:] {
		oldHashes[strings.SplitN(file, "_", 2)[0]] = file // strings.SplitN(file, "_", 2)[1]
	}

	// get a list of all new hashes - Same
	newFiles := getAllFilesIn(newPath)
	for _, file := range newFiles[1:] {
		newHashes[strings.SplitN(file, "_", 2)[0]] = file // strings.SplitN(file, "_", 2)[1]
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

func saveLogs(dir string, newJsFiles []string, allDiffs map[string][]byte) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
	dt := time.Now()
	file, err := os.Create(fmt.Sprintf("%s/log_%s", dir, dt.Format(time.RFC822)))
	defer file.Close()

	w := bufio.NewWriter(file)
	_, err = w.WriteString("New files : \n\n")
	for _, new := range newJsFiles {
		_, err = w.WriteString(fmt.Sprintf("%s\n", new))
	}
	_, err = w.WriteString("\n\nAll diffs :")
	for files, diff := range allDiffs {
		_, err = w.WriteString(fmt.Sprintf("\n\n%s:\n\n", files))
		_, err = w.WriteString(string(diff))
	}
	w.Flush()
}

func main() {
	archiveDir := "./Archives"
	oldDir := "./Downloads"
	newDir := "./temp_Downloads"
	logsDir := "./logs"
	var urlsList []string
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		urlsList = append(urlsList, sc.Text())
	}

	_, err := os.Stat(newDir)
	if os.IsNotExist(err) {
		os.Mkdir(newDir, 0755)
	}

	for _, url := range urlsList {
		downloadAndHash(url)
	}

	allDiffs := make(map[string][]byte)

	newJsFiles, diffJsFiles := checkForNewHashes(oldDir, newDir)
	for _, files := range diffJsFiles {
		allDiffs[fmt.Sprintf("%s & %s", files[0], files[1])] = getDiff(fmt.Sprintf("%s/%s", newDir, files[0]), fmt.Sprintf("%s/%s", oldDir, files[1]))
	}

	fmt.Printf("Downloaded %d files in %s\n", len(urlsList), newDir)
	fmt.Printf("%d new files found\n", len(newJsFiles))
	fmt.Printf("%d files with modifications found\n", len(diffJsFiles))

	_, err = os.Stat(archiveDir)
	if os.IsNotExist(err) {
		os.Mkdir(archiveDir, 0755)
	}

	err = exec.Command("sh", "-c", fmt.Sprintf("mv -n %s/* %s", oldDir, archiveDir)).Run()
	err = exec.Command("sh", "-c", fmt.Sprintf("rm -rf %s; mv %s/ %s", oldDir, newDir, oldDir)).Run()
	if len(newJsFiles) != 0 || len(allDiffs) != 0 {
		saveLogs(logsDir, newJsFiles, allDiffs)
	}
}

// a faire :
// Rajouter notify
// Rajouter les arguments : pour les dossiers, pour notify, pour les output (-s, et -v)
// Creer binary
// upload to git(public)
// publish
// make a help/read-me
