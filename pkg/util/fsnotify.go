package util

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

func StartFilesystemWatchProcess(
	watcher *fsnotify.Watcher,
	dieOnChange bool,
	logger *logrus.Entry,
) {
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					message := fmt.Sprintf("The file '%v' has been modified by another process", event.Name)
					if dieOnChange {
						logger.Fatalln(message + ". Quitting because it may have corrupted data and 'dieOnInputChange' is 'true'.")
					} else {
						logger.Warnln(message + ", but 'dieOnInputChange' is 'false'. This could have unpredictable consequences.")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Errorf("Error while watching file: %v", err)
			}
		}
	}()
}
