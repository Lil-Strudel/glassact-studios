package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrEditConflict means the row changed (or was created) since the caller read
// it, so their write was based on stale state.
var ErrEditConflict = errors.New("edit conflict")

type StoneShape string

type stoneShapes struct {
	FlatTop        StoneShape
	Serpentine     StoneShape
	HalfSerpentine StoneShape
	OvalTop        StoneShape
	Cathedral      StoneShape
	Heart          StoneShape
	FlatMarker     StoneShape
	Slant          StoneShape
}

var StoneShapes = stoneShapes{
	FlatTop:        StoneShape("flat-top"),
	Serpentine:     StoneShape("serpentine"),
	HalfSerpentine: StoneShape("half-serpentine"),
	OvalTop:        StoneShape("oval-top"),
	Cathedral:      StoneShape("cathedral"),
	Heart:          StoneShape("heart"),
	FlatMarker:     StoneShape("flat-marker"),
	Slant:          StoneShape("slant"),
}

func (s StoneShape) IsValid() bool {
	switch s {
	case StoneShapes.FlatTop, StoneShapes.Serpentine, StoneShapes.HalfSerpentine, StoneShapes.OvalTop,
		StoneShapes.Cathedral, StoneShapes.Heart, StoneShapes.FlatMarker, StoneShapes.Slant:
		return true
	}
	return false
}

type StoneElementType string

type stoneElementTypes struct {
	Inlay      StoneElementType
	Background StoneElementType
	Text       StoneElementType
}

var StoneElementTypes = stoneElementTypes{
	Inlay:      StoneElementType("inlay"),
	Background: StoneElementType("background"),
	Text:       StoneElementType("text"),
}

// StoneElement is one thing placed on the stone. X and Y are its center in
// inches from the stone's top-left. Fields after GroupID only apply to the
// element type named in their comment.
type StoneElement struct {
	ID       string           `json:"id"`
	Type     StoneElementType `json:"type"`
	X        float64          `json:"x"`
	Y        float64          `json:"y"`
	Rotation float64          `json:"rotation"`
	GroupID  *string          `json:"group_id"`

	// inlay; its size comes from the inlay itself, never from the layout.
	InlayUUID string `json:"inlay_uuid,omitempty"`

	// background
	BackgroundUUID string  `json:"background_uuid,omitempty"`
	Width          float64 `json:"width,omitempty"`
	FlipX          bool    `json:"flip_x,omitempty"`

	// text
	Text   string  `json:"text,omitempty"`
	Font   string  `json:"font,omitempty"`
	Size   float64 `json:"size,omitempty"`
	Align  string  `json:"align,omitempty"`
	Finish string  `json:"finish,omitempty"`
}

type StoneGroup struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CompoundUUID *string `json:"compound_uuid"`
}

type StoneLayout struct {
	Elements []StoneElement `json:"elements"`
	Groups   []StoneGroup   `json:"groups"`
}

const (
	MaxStoneElements   = 300
	MaxStoneTextLength = 500
	maxStoneInches     = 240
)

var (
	stoneTextFonts   = map[string]bool{"roman": true, "serif": true, "script": true}
	stoneTextAligns  = map[string]bool{"left": true, "center": true, "right": true}
	stoneTextFinish  = map[string]bool{"engraved": true, "frosted": true, "black-fill": true}
	clientIDPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	graniteKeyFormat = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)
	hexColorFormat   = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

// Validate checks everything that can be checked without the database. Whether
// an inlay belongs to the project, or a catalog item is a background, is the
// caller's job.
func (l StoneLayout) Validate() error {
	if len(l.Elements) > MaxStoneElements {
		return fmt.Errorf("a stone can hold at most %d elements", MaxStoneElements)
	}

	groupIDs := make(map[string]bool, len(l.Groups))
	for _, group := range l.Groups {
		if !clientIDPattern.MatchString(group.ID) {
			return fmt.Errorf("invalid group id %q", group.ID)
		}
		if groupIDs[group.ID] {
			return fmt.Errorf("duplicate group id %q", group.ID)
		}
		if len(group.Name) > 255 {
			return fmt.Errorf("group %q name is too long", group.ID)
		}
		groupIDs[group.ID] = true
	}

	elementIDs := make(map[string]bool, len(l.Elements))
	placedInlays := make(map[string]bool)
	for _, el := range l.Elements {
		if !clientIDPattern.MatchString(el.ID) {
			return fmt.Errorf("invalid element id %q", el.ID)
		}
		if elementIDs[el.ID] {
			return fmt.Errorf("duplicate element id %q", el.ID)
		}
		elementIDs[el.ID] = true

		if !isFiniteWithin(el.X, -maxStoneInches, 2*maxStoneInches) ||
			!isFiniteWithin(el.Y, -maxStoneInches, 2*maxStoneInches) ||
			!isFiniteWithin(el.Rotation, -360, 360) {
			return fmt.Errorf("element %q has an out-of-range position or rotation", el.ID)
		}
		if el.GroupID != nil && !groupIDs[*el.GroupID] {
			return fmt.Errorf("element %q references unknown group %q", el.ID, *el.GroupID)
		}

		switch el.Type {
		case StoneElementTypes.Inlay:
			if _, err := uuid.Parse(el.InlayUUID); err != nil {
				return fmt.Errorf("element %q has an invalid inlay_uuid", el.ID)
			}
			if placedInlays[el.InlayUUID] {
				return fmt.Errorf("inlay %s is placed more than once", el.InlayUUID)
			}
			placedInlays[el.InlayUUID] = true
		case StoneElementTypes.Background:
			if _, err := uuid.Parse(el.BackgroundUUID); err != nil {
				return fmt.Errorf("element %q has an invalid background_uuid", el.ID)
			}
			if !isFiniteWithin(el.Width, 0.25, maxStoneInches) {
				return fmt.Errorf("element %q has an out-of-range width", el.ID)
			}
		case StoneElementTypes.Text:
			if len([]rune(el.Text)) > MaxStoneTextLength {
				return fmt.Errorf("element %q text exceeds %d characters", el.ID, MaxStoneTextLength)
			}
			if !stoneTextFonts[el.Font] || !stoneTextAligns[el.Align] || !stoneTextFinish[el.Finish] {
				return fmt.Errorf("element %q has an unknown font, align or finish", el.ID)
			}
			if !isFiniteWithin(el.Size, 0.1, 24) {
				return fmt.Errorf("element %q has an out-of-range text size", el.ID)
			}
		default:
			return fmt.Errorf("element %q has unknown type %q", el.ID, el.Type)
		}
	}

	return nil
}

func isFiniteWithin(v, min, max float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= min && v <= max
}

type StoneDesign struct {
	StandardTable
	ProjectID        int         `json:"project_id"`
	StoneShape       StoneShape  `json:"stone_shape"`
	StoneWidth       float64     `json:"stone_width"`
	StoneHeight      float64     `json:"stone_height"`
	GraniteKey       string      `json:"granite_key"`
	GraniteCustomHex *string     `json:"granite_custom_hex"`
	Layout           StoneLayout `json:"layout"`
}

const DefaultGraniteKey = "georgia-gray"

// NewDefaultStoneDesign is the stone a project starts on when something is
// placed before the dealership has chosen one.
func NewDefaultStoneDesign(projectID int) *StoneDesign {
	return &StoneDesign{
		ProjectID:   projectID,
		StoneShape:  StoneShapes.Serpentine,
		StoneWidth:  36,
		StoneHeight: 24,
		GraniteKey:  DefaultGraniteKey,
		Layout:      StoneLayout{Elements: []StoneElement{}, Groups: []StoneGroup{}},
	}
}

func (d *StoneDesign) Validate() error {
	if !d.StoneShape.IsValid() {
		return fmt.Errorf("unknown stone shape %q", d.StoneShape)
	}
	if !isFiniteWithin(d.StoneWidth, 1, maxStoneInches) || !isFiniteWithin(d.StoneHeight, 1, maxStoneInches) {
		return fmt.Errorf("stone dimensions must be between 1 and %d inches", maxStoneInches)
	}
	if !graniteKeyFormat.MatchString(d.GraniteKey) {
		return fmt.Errorf("invalid granite key %q", d.GraniteKey)
	}
	if d.GraniteCustomHex != nil && !hexColorFormat.MatchString(*d.GraniteCustomHex) {
		return fmt.Errorf("invalid granite color %q", *d.GraniteCustomHex)
	}
	return d.Layout.Validate()
}

type StoneDesignModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func stoneDesignFromGen(gen model.ProjectStoneDesigns) (*StoneDesign, error) {
	design := &StoneDesign{
		StandardTable: StandardTable{
			ID:        int(gen.ID),
			UUID:      gen.UUID.String(),
			CreatedAt: gen.CreatedAt,
			UpdatedAt: gen.UpdatedAt,
			Version:   int(gen.Version),
		},
		ProjectID:        int(gen.ProjectID),
		StoneShape:       StoneShape(gen.StoneShape),
		StoneWidth:       gen.StoneWidth,
		StoneHeight:      gen.StoneHeight,
		GraniteKey:       gen.GraniteKey,
		GraniteCustomHex: gen.GraniteCustomHex,
	}

	if err := json.Unmarshal([]byte(gen.Layout), &design.Layout); err != nil {
		return nil, fmt.Errorf("failed to decode layout of stone design %d: %w", gen.ID, err)
	}
	if design.Layout.Elements == nil {
		design.Layout.Elements = []StoneElement{}
	}
	if design.Layout.Groups == nil {
		design.Layout.Groups = []StoneGroup{}
	}

	return design, nil
}

func stoneDesignToGen(d *StoneDesign) (*model.ProjectStoneDesigns, error) {
	var designUUID uuid.UUID
	if d.UUID != "" {
		parsed, err := uuid.Parse(d.UUID)
		if err != nil {
			return nil, err
		}
		designUUID = parsed
	}

	layout := d.Layout
	if layout.Elements == nil {
		layout.Elements = []StoneElement{}
	}
	if layout.Groups == nil {
		layout.Groups = []StoneGroup{}
	}
	layoutJSON, err := json.Marshal(layout)
	if err != nil {
		return nil, fmt.Errorf("failed to encode stone layout: %w", err)
	}

	return &model.ProjectStoneDesigns{
		ID:               int32(d.ID),
		UUID:             designUUID,
		ProjectID:        int32(d.ProjectID),
		StoneShape:       string(d.StoneShape),
		StoneWidth:       d.StoneWidth,
		StoneHeight:      d.StoneHeight,
		GraniteKey:       d.GraniteKey,
		GraniteCustomHex: d.GraniteCustomHex,
		Layout:           string(layoutJSON),
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
		Version:          int32(d.Version),
	}, nil
}

func (m StoneDesignModel) GetByProjectID(projectID int) (*StoneDesign, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.getByProjectID(ctx, m.STDB, projectID, false)
}

// TxGetByProjectIDForUpdate locks the design row so a read-modify-write inside
// tx cannot interleave with another one.
func (m StoneDesignModel) TxGetByProjectIDForUpdate(tx *sql.Tx, projectID int) (*StoneDesign, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.getByProjectID(ctx, tx, projectID, true)
}

func (m StoneDesignModel) getByProjectID(ctx context.Context, db qrm.Queryable, projectID int, forUpdate bool) (*StoneDesign, bool, error) {
	query := postgres.SELECT(
		table.ProjectStoneDesigns.AllColumns,
	).FROM(
		table.ProjectStoneDesigns,
	).WHERE(
		table.ProjectStoneDesigns.ProjectID.EQ(postgres.Int(int64(projectID))),
	)
	if forUpdate {
		query = query.FOR(postgres.UPDATE())
	}

	var dest model.ProjectStoneDesigns
	err := query.QueryContext(ctx, db, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	design, err := stoneDesignFromGen(dest)
	if err != nil {
		return nil, false, err
	}
	return design, true, nil
}

func (m StoneDesignModel) Insert(design *StoneDesign) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.insert(ctx, m.STDB, design)
}

func (m StoneDesignModel) TxInsert(tx *sql.Tx, design *StoneDesign) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.insert(ctx, tx, design)
}

// insert returns ErrEditConflict when the project already has a design: two
// first saves raced, and the loser must reload rather than silently replace.
func (m StoneDesignModel) insert(ctx context.Context, db qrm.Queryable, design *StoneDesign) error {
	gen, err := stoneDesignToGen(design)
	if err != nil {
		return err
	}

	query := table.ProjectStoneDesigns.INSERT(
		table.ProjectStoneDesigns.ProjectID,
		table.ProjectStoneDesigns.StoneShape,
		table.ProjectStoneDesigns.StoneWidth,
		table.ProjectStoneDesigns.StoneHeight,
		table.ProjectStoneDesigns.GraniteKey,
		table.ProjectStoneDesigns.GraniteCustomHex,
		table.ProjectStoneDesigns.Layout,
	).MODEL(
		gen,
	).ON_CONFLICT(
		table.ProjectStoneDesigns.ProjectID,
	).DO_NOTHING().RETURNING(
		table.ProjectStoneDesigns.ID,
		table.ProjectStoneDesigns.UUID,
		table.ProjectStoneDesigns.CreatedAt,
		table.ProjectStoneDesigns.UpdatedAt,
		table.ProjectStoneDesigns.Version,
	)

	var dest model.ProjectStoneDesigns
	err = query.QueryContext(ctx, db, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}

	design.ID = int(dest.ID)
	design.UUID = dest.UUID.String()
	design.CreatedAt = dest.CreatedAt
	design.UpdatedAt = dest.UpdatedAt
	design.Version = int(dest.Version)
	return nil
}

func (m StoneDesignModel) Update(design *StoneDesign) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.update(ctx, m.STDB, design)
}

func (m StoneDesignModel) TxUpdate(tx *sql.Tx, design *StoneDesign) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return m.update(ctx, tx, design)
}

func (m StoneDesignModel) update(ctx context.Context, db qrm.Queryable, design *StoneDesign) error {
	gen, err := stoneDesignToGen(design)
	if err != nil {
		return err
	}

	query := table.ProjectStoneDesigns.UPDATE(
		table.ProjectStoneDesigns.StoneShape,
		table.ProjectStoneDesigns.StoneWidth,
		table.ProjectStoneDesigns.StoneHeight,
		table.ProjectStoneDesigns.GraniteKey,
		table.ProjectStoneDesigns.GraniteCustomHex,
		table.ProjectStoneDesigns.Layout,
		table.ProjectStoneDesigns.Version,
	).MODEL(
		gen,
	).WHERE(
		postgres.AND(
			table.ProjectStoneDesigns.ID.EQ(postgres.Int(int64(design.ID))),
			table.ProjectStoneDesigns.Version.EQ(postgres.Int(int64(design.Version))),
		),
	).RETURNING(
		table.ProjectStoneDesigns.UpdatedAt,
		table.ProjectStoneDesigns.Version,
	)

	var dest model.ProjectStoneDesigns
	err = query.QueryContext(ctx, db, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}

	design.UpdatedAt = dest.UpdatedAt
	design.Version = int(dest.Version)
	return nil
}

// TxAppendGroup places a group of elements on the project's stone, creating the
// project's design with defaults if it has none yet. Element positions are
// offsets from the stone's center.
func (m StoneDesignModel) TxAppendGroup(tx *sql.Tx, projectID int, group StoneGroup, elements []StoneElement) (*StoneDesign, error) {
	design, found, err := m.TxGetByProjectIDForUpdate(tx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to load stone design for project %d: %w", projectID, err)
	}
	if !found {
		design = NewDefaultStoneDesign(projectID)
	}

	centerX, centerY := design.StoneWidth/2, design.StoneHeight/2
	groupID := group.ID
	design.Layout.Groups = append(design.Layout.Groups, group)
	for _, el := range elements {
		el.X += centerX
		el.Y += centerY
		el.GroupID = &groupID
		design.Layout.Elements = append(design.Layout.Elements, el)
	}

	if err := design.Validate(); err != nil {
		return nil, fmt.Errorf("stone design for project %d would be invalid: %w", projectID, err)
	}

	if found {
		err = m.TxUpdate(tx, design)
	} else {
		err = m.TxInsert(tx, design)
	}
	if err != nil {
		return nil, err
	}
	return design, nil
}

// DuplicateOffsetInches is how far a duplicated inlay lands from its original,
// down and to the right, so the copy is visibly separate rather than hidden
// exactly underneath.
const DuplicateOffsetInches = 1.0

// TxPlaceDuplicate puts copyInlayUUID on the stone next to sourceInlayUUID, with
// the same rotation and ungrouped. It reports false, changing nothing, when the
// project has no design or the source inlay is not on it.
func (m StoneDesignModel) TxPlaceDuplicate(tx *sql.Tx, projectID int, sourceInlayUUID, copyInlayUUID string) (bool, error) {
	design, found, err := m.TxGetByProjectIDForUpdate(tx, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to load stone design for project %d: %w", projectID, err)
	}
	if !found {
		return false, nil
	}

	var source *StoneElement
	for i := range design.Layout.Elements {
		el := &design.Layout.Elements[i]
		if el.Type == StoneElementTypes.Inlay && el.InlayUUID == sourceInlayUUID {
			source = el
			break
		}
	}
	if source == nil {
		return false, nil
	}

	design.Layout.Elements = append(design.Layout.Elements, StoneElement{
		ID:        uuid.NewString(),
		Type:      StoneElementTypes.Inlay,
		X:         source.X + DuplicateOffsetInches,
		Y:         source.Y + DuplicateOffsetInches,
		Rotation:  source.Rotation,
		InlayUUID: copyInlayUUID,
	})

	if err := design.Validate(); err != nil {
		return false, fmt.Errorf("stone design for project %d would be invalid: %w", projectID, err)
	}

	if err := m.TxUpdate(tx, design); err != nil {
		return false, err
	}
	return true, nil
}
