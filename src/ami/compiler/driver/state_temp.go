package driver

import "fmt"

func (s *lowerState) newTemp() string {
    s.temp++
    return fmt.Sprintf("t%d", s.temp)
}

