package amalgam

import (
    "context"
)

func FakeContext() context.Context {
	ctx := context.Background()
	return ctx
}
