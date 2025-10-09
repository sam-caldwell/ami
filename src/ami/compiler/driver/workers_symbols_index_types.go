package driver

type workersSymbolsIndex struct {
    Schema  string `json:"schema"`
    Package string `json:"package"`
    Symbols []struct{
        Name      string `json:"name"`
        Sanitized string `json:"sanitized"`
        Impl      string `json:"impl"`
    } `json:"symbols"`
}

