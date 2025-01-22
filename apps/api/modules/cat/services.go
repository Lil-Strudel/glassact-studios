package cat

import (
	"context"

	"github.com/Lil-Strudel/glassact-studios/apps/api/database"
)

func GetCatsSvc() ([]Cat, error) {
	rows, err := database.Db.Query(context.Background(), `
		SELECT id, name
		FROM cats
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cats []Cat

	for rows.Next() {
		var cat Cat
		err := rows.Scan(&cat.ID, &cat.Name)
		if err != nil {
			return nil, err
		}
		cats = append(cats, cat)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return cats, err
}

func CreateCatSvc(name string) (int, error) {
	var id int
	err := database.Db.QueryRow(context.Background(), `
		INSERT INTO cats (name) 
		VALUES ($1) 
		RETURNING id
	`, name).Scan(&id)

	return id, err
}
