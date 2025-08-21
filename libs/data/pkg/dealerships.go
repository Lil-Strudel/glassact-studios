package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Address struct {
	Street     string  `json:"street"`
	StreetExt  *string `json:"street_ext"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

type Dealership struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Address   Address   `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
}

type DealershipModel struct {
	DB *pgxpool.Pool
}

func (m DealershipModel) Insert(dealership *Dealership) error {
	query := `
        INSERT INTO dealerships (name, street, street_ext, city, state, postal_code, country, location) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, ST_SetSRID(ST_MakePoint($8, $9), 4326)::GEOGRAPHY)
        RETURNING id, uuid, created_at, updated_at, version`

	args := []any{
		dealership.Name,
		dealership.Address.Street,
		dealership.Address.StreetExt,
		dealership.Address.City,
		dealership.Address.State,
		dealership.Address.PostalCode,
		dealership.Address.Country,
		dealership.Address.Longitude,
		dealership.Address.Latitude,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(
		&dealership.ID,
		&dealership.UUID,
		&dealership.CreatedAt,
		&dealership.UpdatedAt,
		&dealership.Version,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m DealershipModel) GetByID(id int) (*Dealership, bool, error) {
	query := `
	SELECT id, uuid, name, street, street_ext, city, state, postal_code, country, ST_X(location::GEOMETRY) AS longitude, ST_Y(location::GEOMETRY) AS latitude, created_at, updated_at, version
        FROM dealerships
        WHERE id = $1`

	var dealership Dealership

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, id).Scan(
		&dealership.ID,
		&dealership.UUID,
		&dealership.Address.Street,
		&dealership.Address.StreetExt,
		&dealership.Address.City,
		&dealership.Address.State,
		&dealership.Address.PostalCode,
		&dealership.Address.Country,
		&dealership.Address.Longitude,
		&dealership.Address.Latitude,
		&dealership.CreatedAt,
		&dealership.UpdatedAt,
		&dealership.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &dealership, true, nil
}

func (m DealershipModel) GetByUUID(uuid string) (*Dealership, bool, error) {
	query := `
	SELECT id, uuid, name, street, street_ext, city, state, postal_code, country, ST_X(location::GEOMETRY) AS longitude, ST_Y(location::GEOMETRY) AS latitude, created_at, updated_at, version
        FROM dealerships
        WHERE uuid = $1`

	var dealership Dealership

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, uuid).Scan(
		&dealership.ID,
		&dealership.UUID,
		&dealership.Address.Street,
		&dealership.Address.StreetExt,
		&dealership.Address.City,
		&dealership.Address.State,
		&dealership.Address.PostalCode,
		&dealership.Address.Country,
		&dealership.Address.Longitude,
		&dealership.Address.Latitude,
		&dealership.CreatedAt,
		&dealership.UpdatedAt,
		&dealership.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &dealership, true, nil
}

func (m DealershipModel) Update(dealership *Dealership) error {
	query := `
        UPDATE dealerships
		SET name = $1, street = $2, street_ext = $3, city = $4, state = $5, postal_code = $6, country = $7, location = ST_SetSRID(ST_MakePoint($8, $9), 4326)::GEOGRAPHY), version = version + 1
		WHERE id = $10 AND version = $11
        RETURNING version`

	args := []any{
		dealership.Name,
		dealership.Address.Street,
		dealership.Address.StreetExt,
		dealership.Address.City,
		dealership.Address.State,
		dealership.Address.PostalCode,
		dealership.Address.Country,
		dealership.Address.Longitude,
		dealership.Address.Latitude,
		dealership.ID,
		dealership.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&dealership.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m DealershipModel) Delete(id int) error {
	query := `
        DELETE FROM dealerships
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
