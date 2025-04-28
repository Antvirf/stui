package model

import "github.com/antvirf/stui/internal/logger"

func IntMax(a, b int) int {
	if a > b {
		logger.Debugf("IntMax: %d > %d, returning %d", a, b, a)
		return a
	}
	logger.Debugf("IntMax: %d <= %d, returning %d", a, b, b)
	return b
}
