package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	nginxMainConf string
	reloadDirs    []string
)

func main() {
	nginxMainConf = os.Getenv("NGINX_MAIN_CONF")
	if nginxMainConf == "" {
		nginxMainConf = "/etc/nginx/nginx.conf"
	}

	reloadDirsStr := os.Getenv("RELOAD_DIRS")
	if reloadDirsStr == "" {
		reloadDirsStr = "/etc/nginx"
	}
	reloadDirs = strings.Split(reloadDirsStr, ",") // 目录由逗号分隔

	watchConfigFiles()
}

func watchConfigFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if isRelevantEvent(event) {
					if err := reloadNginx(); err != nil {
						log.Printf("Failed to reload nginx: %v", err)
					} else {
						log.Println("Nginx reloaded successfully")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()

	for _, dir := range reloadDirs {
		if err := addWatchRecursive(watcher, dir); err != nil {
			log.Printf("Failed to watch directory %s: %v", dir, err)
		}
	}

	<-done
}

func addWatchRecursive(watcher *fsnotify.Watcher, path string) error {
	err := watcher.Add(path)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err := addWatchRecursive(watcher, filepath.Join(path, file.Name()))
			if err != nil {
				log.Printf("Failed to watch subdirectory %s: %v", filepath.Join(path, file.Name()), err)
			}
		}
	}

	return nil
}

func isRelevantEvent(event fsnotify.Event) bool {
	for _, dir := range reloadDirs {
		if strings.HasPrefix(event.Name, dir) {
			return true
		}
	}
	return false
}

func reloadNginx() error {
	cmd := exec.Command("nginx", "-t", "-c", nginxMainConf)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx test config failed: %s", output)
	}

	cmd = exec.Command("killall", "-HUP", "nginx")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx reload failed: %s", output)
	}
	return nil
}
