package dto

type EncryptedMetrics struct {
	Payload string `json:"payload"`
	Secret  string `json:"secret"`
	Nonce   string `json:"nonce"`
}
