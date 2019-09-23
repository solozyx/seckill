package model

// 商品模型
type Product struct {
	ID           int64  `json:"id" sql:"id" form:"ID"`
	ProductName  string `json:"product_mame" sql:"product_name" form:"ProductName"`
	ProductNum   int64  `json:"product_num" sql:"product_num" form:"ProductNum"`
	ProductImage string `json:"product_image" sql:"product_image" form:"ProductImage"`
	ProductUrl   string `json:"product_url" sql:"product_url" form:"ProductUrl"`
}
