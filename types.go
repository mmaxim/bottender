package main

type DrinkID int
type IngredientID int

type Ingredient struct {
	ID       IngredientID
	Name     string
	Category string
	Desc     string
}

type DrinkIngredient struct {
	Ingredient Ingredient
	Amount     int
}

type Drink struct {
	ID          DrinkID
	Ingredients []DrinkIngredient
	Mixing      string
	Serving     string
	Glass       string
	Notes       string
	Name        string
	Author      string
}
