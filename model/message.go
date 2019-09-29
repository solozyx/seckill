package model

type Message struct {
	ProductID int64
	UserID    int64
}

func NewMessage(userId int64, productId int64) *Message {
	return &Message{UserID: userId, ProductID: productId}
}
