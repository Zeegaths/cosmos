package types

type Task struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Creator     string `json:"creator"`
    Bounty      string `json:"bounty"`
    Status      string `json:"status"`
    Claimer     string `json:"claimer,omitempty"`
    Proof       string `json:"proof,omitempty"`
}
