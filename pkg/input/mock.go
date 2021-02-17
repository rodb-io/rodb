package input

type Mock struct {
	data []IterateAllResult
}

func NewMock(data []IterateAllResult) *Mock {
	return &Mock{
		data: data,
	}
}

func (mock *Mock) IterateAll() <-chan IterateAllResult {
	channel := make(chan IterateAllResult)

	go func() {
		defer close(channel)

		for _, row := range mock.data {
			channel <- row
		}
	}()

	return channel
}

func (mock *Mock) Close() error {
	return nil
}
