package conf

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func (p *Conf) Watch(filePath, xDnsPath string, sDnsPath string, reload func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("new watcher error: %s", err)
	}
	go func() {
		var pre time.Time
		defer watcher.Close()
		for {
			select {
			case e := <-watcher.Events:
				if e.Has(fsnotify.Chmod) {
					continue
				}
				if pre.Add(10 * time.Second).After(time.Now()) {
					continue
				}
				pre = time.Now()
				go func() {
					time.Sleep(5 * time.Second)
					switch filepath.Base(strings.TrimSuffix(e.Name, "~")) {
					case filepath.Base(xDnsPath), filepath.Base(sDnsPath):
						log.Println("DNS file changed, reloading...")
					default:
						log.Println("config dir changed, reloading...")
					}
					*p = *New()
					err := p.LoadFromPath(filePath)
					if err != nil {
						log.Printf("reload config error: %s", err)
					}
					reload()
					log.Println("reload config success")
				}()
			case err := <-watcher.Errors:
				if err != nil {
					log.Printf("File watcher error: %s", err)
				}
			}
		}
	}()
	err = watcher.Add(path.Dir(filePath))
	if err != nil {
		return fmt.Errorf("watch file error: %s", err)
	}
	if xDnsPath != "" {
		err = watcher.Add(path.Dir(xDnsPath))
		if err != nil {
			return fmt.Errorf("watch dns file error: %s", err)
		}
	}
	if sDnsPath != "" {
		err = watcher.Add(path.Dir(sDnsPath))
		if err != nil {
			return fmt.Errorf("watch dns file error: %s", err)
		}
	}
	return nil
}
