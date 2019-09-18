package main

import (
	"database/sql"
	"errors"
)

var errDrinkNotFound = errors.New("drink not found")

type DrinkDB struct {
	db *sql.DB
}

func NewDrinkDB(db *sql.DB) *DrinkDB {
	return &DrinkDB{
		db: db,
	}
}

func (d *DrinkDB) Describe(query string) (res Drink, err error) {
	rows, err := d.db.Query(`
		SELECT id, name, mixing, glass, serving, notes
		FROM drinks 
		WHERE name = ?
	`, query)
	if err != nil {
		return res, err
	}

	// Get drink basic stats
	var drinkID int
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&drinkID, &res.Name, &res.Mixing, &res.Glass, &res.Serving, &res.Notes); err != nil {
			return res, err
		}
		break
	}
	rows.Close()
	if len(res.Mixing) == 0 {
		// not found
		return res, errDrinkNotFound
	}

	// Get ingredients
	rows, err = d.db.Query(`
		SELECT i.name, category, i.desc, amount
		FROM drinks AS d
		JOIN drink_ingredients AS di ON d.id=di.drink_id
		JOIN ingredient AS i ON di.ingredient_id=i.id
		WHERE d.id = ?
		ORDER by d.name, amount desc
	`, drinkID)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var ingredient DrinkIngredient
		if err := rows.Scan(&ingredient.Ingredient.Name, &ingredient.Ingredient.Category,
			&ingredient.Ingredient.Desc, &ingredient.Amount); err != nil {
			return res, err
		}
		res.Ingredients = append(res.Ingredients, ingredient)
	}
	return res, nil
}
