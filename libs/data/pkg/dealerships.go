package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Address struct {
	Street     string  `json:"street"`
	StreetExt  string  `json:"street_ext"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

type Dealership struct {
	StandardTable
	Name    string  `json:"name"`
	Address Address `json:"address"`
}

type DealershipModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func dealershipFromGen(genDeal model.Dealerships, longitude, latitude float64) *Dealership {
	dealership := Dealership{
		StandardTable: StandardTable{
			ID:        int(genDeal.ID),
			UUID:      genDeal.UUID.String(),
			CreatedAt: genDeal.CreatedAt,
			UpdatedAt: genDeal.UpdatedAt,
			Version:   int(genDeal.Version),
		},
		Name: genDeal.Name,
		Address: Address{
			Street:     genDeal.Street,
			StreetExt:  genDeal.StreetExt,
			City:       genDeal.City,
			State:      genDeal.State,
			PostalCode: genDeal.PostalCode,
			Country:    genDeal.Country,
			Longitude:  longitude,
			Latitude:   latitude,
		},
	}

	return &dealership
}

func dealershipToGen(d *Dealership) (*model.Dealerships, error) {
	var dealershipUUID uuid.UUID
	var err error

	if d.UUID != "" {
		dealershipUUID, err = uuid.Parse(d.UUID)
		if err != nil {
			return nil, err
		}
	}

	genDeal := model.Dealerships{
		ID:         int32(d.ID),
		UUID:       dealershipUUID,
		Name:       d.Name,
		Street:     d.Address.Street,
		StreetExt:  d.Address.StreetExt,
		City:       d.Address.City,
		State:      d.Address.State,
		PostalCode: d.Address.PostalCode,
		Country:    d.Address.Country,
		UpdatedAt:  d.UpdatedAt,
		CreatedAt:  d.CreatedAt,
		Version:    int32(d.Version),
	}

	return &genDeal, nil
}

func (m DealershipModel) Insert(dealership *Dealership) error {
	genDeal, err := dealershipToGen(dealership)
	if err != nil {
		return err
	}

	// Note: We use VALUES instead of MODEL because the location field requires
	// a custom SQL expression (ST_SetSRID/ST_MakePoint)
	locationExpr := postgres.RawString(
		"ST_SetSRID(ST_MakePoint(#1, #2), 4326)::GEOGRAPHY",
		map[string]interface{}{
			"#1": dealership.Address.Longitude,
			"#2": dealership.Address.Latitude,
		},
	)

	query := table.Dealerships.INSERT(
		table.Dealerships.Name,
		table.Dealerships.Street,
		table.Dealerships.StreetExt,
		table.Dealerships.City,
		table.Dealerships.State,
		table.Dealerships.PostalCode,
		table.Dealerships.Country,
		table.Dealerships.Location,
	).VALUES(
		genDeal.Name,
		genDeal.Street,
		genDeal.StreetExt,
		genDeal.City,
		genDeal.State,
		genDeal.PostalCode,
		genDeal.Country,
		locationExpr,
	).RETURNING(
		table.Dealerships.ID,
		table.Dealerships.UUID,
		table.Dealerships.UpdatedAt,
		table.Dealerships.CreatedAt,
		table.Dealerships.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Dealerships
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	dealership.ID = int(dest.ID)
	dealership.UUID = dest.UUID.String()
	dealership.UpdatedAt = dest.UpdatedAt
	dealership.CreatedAt = dest.CreatedAt
	dealership.Version = int(dest.Version)

	return nil
}

func (m DealershipModel) GetByID(id int) (*Dealership, bool, error) {
	longitudeExpr := postgres.RawString("ST_X(location::GEOMETRY)")
	latitudeExpr := postgres.RawString("ST_Y(location::GEOMETRY)")

	query := postgres.SELECT(
		table.Dealerships.AllColumns,
		longitudeExpr.AS("longitude"),
		latitudeExpr.AS("latitude"),
	).FROM(
		table.Dealerships,
	).WHERE(
		table.Dealerships.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest struct {
		model.Dealerships
		Longitude float64
		Latitude  float64
	}
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return dealershipFromGen(dest.Dealerships, dest.Longitude, dest.Latitude), true, nil
}

func (m DealershipModel) GetByUUID(uuidStr string) (*Dealership, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	longitudeExpr := postgres.RawString("ST_X(location::GEOMETRY)")
	latitudeExpr := postgres.RawString("ST_Y(location::GEOMETRY)")

	query := postgres.SELECT(
		table.Dealerships.AllColumns,
		longitudeExpr.AS("longitude"),
		latitudeExpr.AS("latitude"),
	).FROM(
		table.Dealerships,
	).WHERE(
		table.Dealerships.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest struct {
		model.Dealerships
		Longitude float64
		Latitude  float64
	}
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return dealershipFromGen(dest.Dealerships, dest.Longitude, dest.Latitude), true, nil
}

func (m DealershipModel) GetAll() ([]*Dealership, error) {
	longitudeExpr := postgres.RawString("ST_X(location::GEOMETRY)")
	latitudeExpr := postgres.RawString("ST_Y(location::GEOMETRY)")

	query := postgres.SELECT(
		table.Dealerships.AllColumns,
		longitudeExpr.AS("longitude"),
		latitudeExpr.AS("latitude"),
	).FROM(
		table.Dealerships,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []struct {
		model.Dealerships
		Longitude float64
		Latitude  float64
	}
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	dealerships := make([]*Dealership, len(dest))
	for i, d := range dest {
		dealerships[i] = dealershipFromGen(d.Dealerships, d.Longitude, d.Latitude)
	}

	return dealerships, nil
}

func (m DealershipModel) Update(dealership *Dealership) error {
	genDeal, err := dealershipToGen(dealership)
	if err != nil {
		return err
	}

	locationExpr := postgres.RawString(
		"ST_SetSRID(ST_MakePoint(#1, #2), 4326)::GEOGRAPHY",
		map[string]interface{}{
			"#1": dealership.Address.Longitude,
			"#2": dealership.Address.Latitude,
		},
	)

	query := table.Dealerships.UPDATE(
		table.Dealerships.Name,
		table.Dealerships.Street,
		table.Dealerships.StreetExt,
		table.Dealerships.City,
		table.Dealerships.State,
		table.Dealerships.PostalCode,
		table.Dealerships.Country,
		table.Dealerships.Location,
		table.Dealerships.Version,
	).MODEL(
		genDeal,
	).SET(
		table.Dealerships.Location.SET(locationExpr),
	).WHERE(
		postgres.AND(
			table.Dealerships.ID.EQ(postgres.Int(int64(dealership.ID))),
			table.Dealerships.Version.EQ(postgres.Int(int64(dealership.Version))),
		),
	).RETURNING(
		table.Dealerships.UpdatedAt,
		table.Dealerships.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Dealerships
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	dealership.UpdatedAt = dest.UpdatedAt
	dealership.Version = int(dest.Version)

	return nil
}

func (m DealershipModel) Delete(id int) error {
	query := table.Dealerships.DELETE().WHERE(
		table.Dealerships.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
