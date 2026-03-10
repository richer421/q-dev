package model

type HelloWorld struct {
	BaseModel
	Title       string `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description string `gorm:"column:description;type:text" json:"description"`
}
