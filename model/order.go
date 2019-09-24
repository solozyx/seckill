package model

type Order struct {
	ID          int64 `sql:"id"`
	UserId      int64 `sql:"user_id"`
	ProductId   int64 `sql:"product_id"`
	OrderStatus int   `sql:"order_status"`
}

const (
	OrderWait    = iota // 0 初始值
	OrderSuccess        // 1
	OrderFailed         // 2
)
