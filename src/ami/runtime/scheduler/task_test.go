package scheduler

import (
    "context"
    "testing"
)

func TestTask_Empty(t *testing.T) {
    _ = Task{Do: func(context.Context){}}
}

