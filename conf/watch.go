package conf

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"path"
	"time"
)

func (p *Conf) Watch(filePath, dnsPath string, reload func()) error {
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
					time.Sleep(10 * time.Second)
					if e.Name == dnsPath {
						log.Println("DNS file changed, reloading...")
					} else {
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
	if dnsPath != "" {
		err = watcher.Add(path.Dir(dnsPath))
		if err != nil {
			return fmt.Errorf("watch dns file error: %s", err)
		}
	}
	return nil
}
