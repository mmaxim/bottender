package main

type Ingredient struct {
	Name     string
	Category string
	Desc     string
}

type DrinkIngredient struct {
	Ingredient Ingredient
	Amount     int
}

type Drink struct {
	Ingredients []DrinkIngredient
	Mixing      string
	Serving     string
	Glass       string
	Notes       string
	Name        string
}
