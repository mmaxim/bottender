package main

import (
	"fmt"
	"strings"
)

var mdQuotes = "```"

func (i DrinkIngredient) DisplayFull() string {
	var amount string
	switch i.Ingredient.Category {
	case "spirit", "liqueur", "aromatic", "sugar", "citrus":
		amount = fmt.Sprintf("%.2f oz", float64(i.Amount)/100.0)
	case "bitters":
		amount = fmt.Sprintf("%d dashes", i.Amount)
	case "garnish":
		amount = fmt.Sprintf("%d", i.Amount)
	}
	return " " + amount + " " + i.Ingredient.Name
}

func DisplayDrinkFull(drink Drink) string {
	var ingredients []string
	for _, i := range drink.Ingredients {
		ingredients = append(ingredients, i.DisplayFull())
	}
	return fmt.Sprintf(`*Name*: %s
*Mixing*: %s
*Glass*: %s
*Serving*: %s
*Ingredients*
%s
*Notes*
 %s`, drink.Name, drink.Mixing, drink.Glass, drink.Serving, strings.Join(ingredients, "\n"),
		drink.Notes)
}
