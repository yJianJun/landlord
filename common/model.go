package common

// Account代表一个用户帐户，具有以下属性：
// - Id: 代表账户ID的整数
// - Email: 代表帐户电子邮件的字符串
// - 用户名：代表帐户用户名的字符串
// - 密码：代表账户密码的字符串
// - Coin: 代表账户币余额的整数
// - CreatedDate：表示帐户创建日期的字符串
// - UpdateDate：代表帐户上次更新日期的字符串
type Account struct {
	Id          int    `json:"id" db:"id"`
	Email       string `json:"email" db:"email"`
	Username    string `json:"username" db:"username"`
	Password    string `json:"password" db:"password"`
	Coin        int    `json:"coin" db:"coin"`
	CreatedDate string `json:"created_date" db:"created_date"`
	UpdateDate  string `json:"updated_date" db:"updated_date"`
}
