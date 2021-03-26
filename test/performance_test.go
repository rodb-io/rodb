package test

import (
	"testing"
)

const iteratorTestIterations = 1000000

func iteratorGeneratorFunctionForTests() func() (int, bool) {
	i := 0
	return func() (int, bool) {
		for i < iteratorTestIterations {
			value := i
			i++
			return value, false
		}

		return 0, true
	}
}
func iteratorCallbackForTests(callback func(value int) bool) {
	for i := 0; i < iteratorTestIterations; i++ {
		shouldContinue := callback(i)
		if !shouldContinue {
			break
		}
	}
}
func iteratorChannelForTests() chan int {
	channel := make(chan int)
	go (func() {
		for i := 0; i < iteratorTestIterations; i++ {
			channel <- i
		}

		close(channel)
	})()

	return channel
}
func iteratorArrayForTests() []int {
	array := make([]int, iteratorTestIterations)
	for i := 0; i < iteratorTestIterations; i++ {
		array[i] = i
	}

	return array
}

func BenchmarkIteratorMethods(b *testing.B) {
	b.Run("generator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iterator := iteratorGeneratorFunctionForTests()
			for {
				value, end := iterator()
				if end {
					break
				}
				_ = value
			}
		}
	})
	b.Run("callback", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iteratorCallbackForTests(func(value int) bool {
				_ = value
				return true
			})
		}
	})
	b.Run("channel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for value := range iteratorChannelForTests() {
				_ = value
			}
		}
	})
	b.Run("array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for value := range iteratorArrayForTests() {
				_ = value
			}
		}
	})
}
