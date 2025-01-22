package cat

type Cat struct {
	ID   int    `json:"id"`
	Name string `json:"name" validate:"required" body:"name`
}
