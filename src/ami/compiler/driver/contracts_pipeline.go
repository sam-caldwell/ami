package driver

type contractPipeline struct {
    Name  string         `json:"name"`
    Steps []contractStep `json:"steps"`
}

