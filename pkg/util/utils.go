package util

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
)

func RemoveCharacters(value string, charactersToRemove string) string {
	for _, c := range charactersToRemove {
		value = strings.ReplaceAll(value, string(c), "")
	}

	return value
}

func IsInArray(value string, array []string) bool {
	for _, arrayElement := range array {
		if arrayElement == value {
			return true
		}
	}

	return false
}

func GetAddress(address net.Addr) string {
	result := address.String()
	for from, to := range map[string]string{
		"[::]:":    "127.0.0.1:",
		"0.0.0.0:": "127.0.0.1:",
	} {
		if strings.HasPrefix(result, from) {
			result = to + result[len(from):]
		}
	}

	return result
}

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
