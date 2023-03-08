package mode

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmorenohj/sports/common/config/db"
	env "github.com/jmorenohj/sports/common/config/envs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Route(app *fiber.App) {

	mode := app.Group("/mode")

	modesCollection := db.Client.Database(env.EnvVariable("CUR_DB")).Collection("modes")
	sportsCollection := db.Client.Database(env.EnvVariable("CUR_DB")).Collection("sports")

	mode.Get("/:sportId", func(c *fiber.Ctx) error {
		sportId := c.Params("sportId")
		objectId, err := primitive.ObjectIDFromHex(sportId)
		query, err := modesCollection.Find(context.TODO(), bson.D{{"sportId", objectId}})
		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No se encontraron los modos de juego del deporte buscado"})
		}

		var modes []bson.M
		if err = query.All(context.TODO(), &modes); err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error recibiendo los datos desde la base de datos"})
		}

		return c.Status(fiber.StatusOK).JSON(modes)

	})

	mode.Post("/", func(c *fiber.Ctx) error {
		mode := new(Mode)

		if err := c.BodyParser(mode); err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al recibir el modo en el body"})
		}
		fmt.Println(mode)

		errors := ValidateMode(*mode)
		if errors != "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errors})
		}

		sportId := mode.SportId

		results, err := modesCollection.InsertOne(context.TODO(), mode)
		_ = results
		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al guardar el modo en base de datos"})
		}
		result, err := sportsCollection.UpdateOne(context.TODO(), bson.D{{"_id", sportId}}, bson.D{{"$push", bson.D{{"modes", results.InsertedID}}}})
		_ = result
		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error al recibir al actualizar el deporte correspondiente"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Modo añadido exitosamente"})
	})

}
