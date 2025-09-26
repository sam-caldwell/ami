package tester

import "reflect"

func deepEqualJSON(a, b any) bool {
    return reflect.DeepEqual(a, b)
}

