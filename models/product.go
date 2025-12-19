package models

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name" validate:"required,min=3"`
	Price int    `json:"price" validate:"required,gt=0"`
	Stock int    `json:"stock" validate:"required,gte=0"`
}

type ProductFilter struct {
	Page   int    `json:"page" validate:"gte=1"`
	Limit  int    `json:"limit" validate:"gte=1,lte=100"`
	Search string `json:"search"`
}

func (f *ProductFilter) GetOffset() int {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 10
	}
	return (f.Page - 1) * f.Limit
}
