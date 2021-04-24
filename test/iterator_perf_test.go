package test

import (
	"testing"
)

const iteratorBenchmarkIterations = 1000000

func generatorFunctionForTests() func() (int, bool) {
	i := 0
	return func() (int, bool) {
		for i < iteratorBenchmarkIterations {
			value := i
			i++
			return value, false
		}

		return 0, true
	}
}
func callbackIteratorForTests(callback func(value int) bool) {
	for i := 0; i < iteratorBenchmarkIterations; i++ {
		shouldContinue := callback(i)
		if !shouldContinue {
			break
		}
	}
}
func channelIteratorForTests() chan int {
	channel := make(chan int)
	go (func() {
		for i := 0; i < iteratorBenchmarkIterations; i++ {
			channel <- i
		}

		close(channel)
	})()

	return channel
}

func BenchmarkIteratorMethods(b *testing.B) {
	b.Run("generator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			next := generatorFunctionForTests()
			for {
				value, end := next()
				if end {
					break
				}
				_ = value
			}
		}
	})
	b.Run("callback", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			callbackIteratorForTests(func(value int) bool {
				_ = value
				return true
			})
		}
	})
	b.Run("channel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for value := range channelIteratorForTests() {
				_ = value
			}
		}
	})
}
