package util

import (
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

type Sizeable interface {
	Size() (int64, error)
}

func TrackProgress(
	sizeable Sizeable,
	logger *logrus.Entry,
) (updateProgress func(position int64)) {
	totalSize, err := sizeable.Size()
	if err != nil {
		logger.Errorf("Cannot determine the total size: '%+v'. The progress will not be displayed.", err)
	} else if totalSize == 0 {
		logger.Infoln("The total size is unknown. The progress will not be displayed.")
	}

	nextProgress := time.Now()

	return func(position int64) {
		if totalSize != 0 {
			if now := time.Now(); now.After(nextProgress) {
				progress := float64(position) / float64(totalSize)
				logger.Infof("Indexing progress: %d%%", int(math.Floor(progress*100)))
				nextProgress = now.Add(time.Second)
			}
		}
	}
}
