package cat

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/model"
)

func GetCats(w http.ResponseWriter, req *http.Request) {
	cats, err := GetCatsSvc()
	if err != nil {
		fmt.Println(err)
		panic("OMG AN ERROR")
	}

	json.NewEncoder(w).Encode(cats)
}

func PostCat(w http.ResponseWriter, req *http.Request) {
	cat := new(model.Cat)

	if err := json.NewDecoder(req.Body).Decode(&cat); err != nil {
		fmt.Println(err)
		panic("OMG AN ERROR")
	}

	id, err := CreateCatSvc(cat.Name)
	if err != nil {
		fmt.Println(err)
		panic("OMG AN ERROR")
	}

	type Response struct {
		Id int `json:"id"`
	}
	json.NewEncoder(w).Encode(Response{
		Id: id,
	})
}
