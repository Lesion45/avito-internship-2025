package entity

type User struct {
	Username string `db:"username"`
	Password []byte `db:"password"`
}
