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

func (d *DrinkDB) runTxn(fn func(tx *sql.Tx) error) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (d *DrinkDB) describeDrinkByID(drinkID DrinkID) (res Drink, err error) {
	res.ID = drinkID
	rows, err := d.db.Query(`
		SELECT name, mixing, glass, serving, notes, author
		FROM drinks
		WHERE id = ?
	`, drinkID)
	if err != nil {
		return res, err
	}

	// Get drink basic stats
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&res.Name, &res.Mixing, &res.Glass, &res.Serving, &res.Notes, &res.Author); err != nil {
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
	defer rows.Close()
	var drinkID DrinkID
	for rows.Next() {
		if err := rows.Scan(&drinkID); err != nil {
			return res, err
		}
		break
	}
	rows.Close()
	return d.describeDrinkByID(drinkID)
}

func (d *DrinkDB) DescribeIngredient(query string) (res Ingredient, err error) {
	rows, err := d.db.Query(`
		SELECT i.id, i.name, i.desc, i.category
		FROM ingredient i
		WHERE name = ?
	`, query)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		if err := rows.Scan(&res.ID, &res.Name, &res.Desc, &res.Category); err != nil {
			return res, err
		}
		found = true
		break
	}
	if !found {
		return res, errors.New("ingredient not found")
	}
	return res, nil
}

func (d *DrinkDB) addDrinkIngredient(tx *sql.Tx, drinkID DrinkID, ingredient DrinkIngredient) error {
	_, err := tx.Exec(`
		INSERT INTO drink_ingredients (drink_id, ingredient_id, amount)
		VALUES (?, ?, ?)
	`, drinkID, ingredient.Ingredient.ID, ingredient.Amount)
	return err
}

func (d *DrinkDB) AddRecipe(name, mixing, glass, serving, notes string, ingredients []DrinkIngredient,
	author string) (err error) {
	return d.runTxn(func(tx *sql.Tx) error {
		nameRes, err := tx.Exec(`
			INSERT INTO drinks (name, mixing, glass, serving, notes, author)
			VALUES (?, ?, ?, ?, ?, ?)
		`, name, mixing, glass, serving, notes, author)
		if err != nil {
			return err
		}
		id, err := nameRes.LastInsertId()
		if err != nil {
			return err
		}
		for _, ingredient := range ingredients {
			if err := d.addDrinkIngredient(tx, DrinkID(id), ingredient); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *DrinkDB) Random(query *string, num int) (res []Drink, err error) {
	var rows *sql.Rows
	if query == nil {
		rows, err = d.db.Query(`SELECT id FROM drinks ORDER BY RAND() LIMIT ?`, num)
	} else {
		rows, err = d.db.Query(`
		SELECT drink_id
		FROM drink_ingredients di
		JOIN (
			SELECT id FROM ingredient WHERE name LIKE ?
		) AS matches ON di.ingredient_id = matches.id
		ORDER BY RAND()
		LIMIT ?
	`, fmt.Sprintf("%%%s%%", *query), num)
	}
	if err != nil {
		return res, err
	}
	defer rows.Close()
	var drinkIDs []DrinkID
	for rows.Next() {
		var drinkID DrinkID
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
