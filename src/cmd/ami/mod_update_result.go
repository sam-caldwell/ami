package main

type modUpdateResult struct {
    Updated  []modUpdateItem `json:"updated"`
    Message  string          `json:"message,omitempty"`
    Audit    *modAuditEmbed  `json:"audit,omitempty"`
    Selected []modUpdateItem `json:"selected,omitempty"`
}

