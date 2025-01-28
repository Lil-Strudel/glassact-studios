package cat

import (
	"fmt"

	"github.com/Lil-Strudel/glassact-studios/apps/api/model"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
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
	sess := session.FromContext(c)

	key, ok := sess.Get("uid").(string)
	if ok {
		fmt.Println(sess.ID(), key)
	} else {
		fmt.Println(sess.ID(), "not ok")
	}

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
