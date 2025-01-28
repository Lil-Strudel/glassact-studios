package cat

import (
	"fmt"

	"github.com/Lil-Strudel/glassact-studios/apps/api/model"
	"github.com/gofiber/fiber/v3"
)

func GetCats(c fiber.Ctx) error {
	cats, err := GetCatsSvc()
	if err != nil {
		fmt.Println(err)
		panic("OMG AN ERROR")
	}

	return c.JSON(cats)
}

func PostCat(c fiber.Ctx) error {
	cat := new(model.Cat)

	if err := c.Bind().Body(cat); err != nil {
		return err
	}

	id, err := CreateCatSvc(cat.Name)
	if err != nil {
		fmt.Println(err)
		panic("OMG AN ERROR")
	}

	return c.JSON(fiber.Map{"id": id})
}
