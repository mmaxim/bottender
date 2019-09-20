package main

import (
	"database/sql"
	"errors"
	"fmt"
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

func (d *DrinkDB) debug(msg string, args ...interface{}) {
	fmt.Printf("DrinkDB: "+msg+"\n", args...)
}

func (d *DrinkDB) describeDrinkByID(drinkID int) (res Drink, err error) {
	rows, err := d.db.Query(`
		SELECT name, mixing, glass, serving, notes
		FROM drinks
		WHERE id = ?
	`, drinkID)
	if err != nil {
		return res, err
	}

	// Get drink basic stats
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&res.Name, &res.Mixing, &res.Glass, &res.Serving, &res.Notes); err != nil {
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

func (d *DrinkDB) Describe(query string) (res Drink, err error) {
	rows, err := d.db.Query(`SELECT id FROM drinks WHERE name = ?`, query)
	if err != nil {
		return res, err
	}
	var drinkID int
	for rows.Next() {
		if err := rows.Scan(&drinkID); err != nil {
			return res, err
		}
		break
	}
	rows.Close()
	return d.describeDrinkByID(drinkID)
}

func (d *DrinkDB) Random(query string, num int) (res []Drink, err error) {
	rows, err := d.db.Query(`
		SELECT drink_id
		FROM drink_ingredients di
		JOIN (
			SELECT id FROM ingredient WHERE name LIKE ?
		) AS matches ON di.ingredient_id = matches.id
		ORDER BY RAND()
		LIMIT ?
	`, fmt.Sprintf("%%%s%%", query), num)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	var drinkIDs []int
	for rows.Next() {
		var drinkID int
		if err := rows.Scan(&drinkID); err != nil {
			return res, err
		}
		drinkIDs = append(drinkIDs, drinkID)
	}
	rows.Close()

	for _, drinkID := range drinkIDs {
		drink, err := d.describeDrinkByID(drinkID)
		if err != nil {
			d.debug("failed to descibe random drink, skipping: %s", err)
			continue
		}
		res = append(res, drink)
	}
	return res, nil
}
