package types

const (
    ROLE_ADMIN = "ADMIN"
    ROLE_USER  = "USER"
)

type User struct {
    Address string `json:"address"`
    Role    string `json:"role"`
}
