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

type SupportCategory string

type supportCategories struct {
	Installation SupportCategory
	Ordering     SupportCategory
	Pricing      SupportCategory
	Contact      SupportCategory
	General      SupportCategory
}

var SupportCategories = supportCategories{
	Installation: SupportCategory("installation"),
	Ordering:     SupportCategory("ordering"),
	Pricing:      SupportCategory("pricing"),
	Contact:      SupportCategory("contact"),
	General:      SupportCategory("general"),
}

type SupportArticle struct {
	StandardTable
	Category    SupportCategory `json:"category"`
	Title       string          `json:"title"`
	Body        string          `json:"body"`
	YoutubeURL  *string         `json:"youtube_url"`
	SortOrder   int             `json:"sort_order"`
	IsPublished bool            `json:"is_published"`
}

type SupportArticleModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func supportArticleFromGen(gen model.SupportArticles) *SupportArticle {
	return &SupportArticle{
		StandardTable: StandardTable{
			ID:        int(gen.ID),
			UUID:      gen.UUID.String(),
			CreatedAt: gen.CreatedAt,
			UpdatedAt: gen.UpdatedAt,
			Version:   int(gen.Version),
		},
		Category:    SupportCategory(gen.Category),
		Title:       gen.Title,
		Body:        gen.Body,
		YoutubeURL:  gen.YoutubeURL,
		SortOrder:   int(gen.SortOrder),
		IsPublished: gen.IsPublished,
	}
}

func supportArticleToGen(a *SupportArticle) (*model.SupportArticles, error) {
	var articleUUID uuid.UUID
	var err error

	if a.UUID != "" {
		articleUUID, err = uuid.Parse(a.UUID)
		if err != nil {
			return nil, err
		}
	}

	return &model.SupportArticles{
		ID:          int32(a.ID),
		UUID:        articleUUID,
		Category:    string(a.Category),
		Title:       a.Title,
		Body:        a.Body,
		YoutubeURL:  a.YoutubeURL,
		SortOrder:   int32(a.SortOrder),
		IsPublished: a.IsPublished,
		UpdatedAt:   a.UpdatedAt,
		CreatedAt:   a.CreatedAt,
		Version:     int32(a.Version),
	}, nil
}

func (m SupportArticleModel) Insert(article *SupportArticle) error {
	gen, err := supportArticleToGen(article)
	if err != nil {
		return err
	}

	query := table.SupportArticles.INSERT(
		table.SupportArticles.Category,
		table.SupportArticles.Title,
		table.SupportArticles.Body,
		table.SupportArticles.YoutubeURL,
		table.SupportArticles.SortOrder,
		table.SupportArticles.IsPublished,
	).MODEL(
		gen,
	).RETURNING(
		table.SupportArticles.ID,
		table.SupportArticles.UUID,
		table.SupportArticles.UpdatedAt,
		table.SupportArticles.CreatedAt,
		table.SupportArticles.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.SupportArticles
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	article.ID = int(dest.ID)
	article.UUID = dest.UUID.String()
	article.UpdatedAt = dest.UpdatedAt
	article.CreatedAt = dest.CreatedAt
	article.Version = int(dest.Version)

	return nil
}

func (m SupportArticleModel) GetByID(id int) (*SupportArticle, bool, error) {
	query := postgres.SELECT(
		table.SupportArticles.AllColumns,
	).FROM(
		table.SupportArticles,
	).WHERE(
		table.SupportArticles.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.SupportArticles
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return supportArticleFromGen(dest), true, nil
}

func (m SupportArticleModel) GetByUUID(uuidStr string) (*SupportArticle, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.SupportArticles.AllColumns,
	).FROM(
		table.SupportArticles,
	).WHERE(
		table.SupportArticles.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.SupportArticles
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return supportArticleFromGen(dest), true, nil
}

func (m SupportArticleModel) GetAll() ([]*SupportArticle, error) {
	query := postgres.SELECT(
		table.SupportArticles.AllColumns,
	).FROM(
		table.SupportArticles,
	).ORDER_BY(
		table.SupportArticles.Category.ASC(),
		table.SupportArticles.SortOrder.ASC(),
		table.SupportArticles.CreatedAt.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.SupportArticles
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	articles := make([]*SupportArticle, len(dest))
	for i, d := range dest {
		articles[i] = supportArticleFromGen(d)
	}

	return articles, nil
}

func (m SupportArticleModel) GetAllPublished() ([]*SupportArticle, error) {
	query := postgres.SELECT(
		table.SupportArticles.AllColumns,
	).FROM(
		table.SupportArticles,
	).WHERE(
		table.SupportArticles.IsPublished.EQ(postgres.Bool(true)),
	).ORDER_BY(
		table.SupportArticles.Category.ASC(),
		table.SupportArticles.SortOrder.ASC(),
		table.SupportArticles.CreatedAt.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.SupportArticles
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	articles := make([]*SupportArticle, len(dest))
	for i, d := range dest {
		articles[i] = supportArticleFromGen(d)
	}

	return articles, nil
}

func (m SupportArticleModel) Update(article *SupportArticle) error {
	gen, err := supportArticleToGen(article)
	if err != nil {
		return err
	}

	query := table.SupportArticles.UPDATE(
		table.SupportArticles.Category,
		table.SupportArticles.Title,
		table.SupportArticles.Body,
		table.SupportArticles.YoutubeURL,
		table.SupportArticles.SortOrder,
		table.SupportArticles.IsPublished,
		table.SupportArticles.Version,
	).MODEL(
		gen,
	).WHERE(
		postgres.AND(
			table.SupportArticles.ID.EQ(postgres.Int(int64(article.ID))),
			table.SupportArticles.Version.EQ(postgres.Int(int64(article.Version))),
		),
	).RETURNING(
		table.SupportArticles.UpdatedAt,
		table.SupportArticles.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.SupportArticles
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	article.UpdatedAt = dest.UpdatedAt
	article.Version = int(dest.Version)

	return nil
}

func (m SupportArticleModel) Delete(id int) error {
	query := table.SupportArticles.DELETE().WHERE(
		table.SupportArticles.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
