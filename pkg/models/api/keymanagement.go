package api

type KeyManagementKeyList struct {
	Keys []string `json:"keys"`
}
type KeyManagementKey struct {
	Private    string `json:"private"`
	Public     string `json:"public"`
	Passphrase string `json:"passphrase"`
}
