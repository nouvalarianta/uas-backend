package config

import (
	"database/sql"
	"os"
	repository "uas-backend/app/repository"
	service "uas-backend/app/service"
	route "uas-backend/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewApp(mDB *mongo.Database, pgDB *sql.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Output: os.Stdout,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Repositories
	userRepo := repository.NewUserRepository(pgDB)
	achievementRepo := repository.NewAchievementRepository(mDB)
	achievementReferenceRepo := repository.NewAchievementReferenceRepository(pgDB)
	studentRepo := repository.NewStudentRepository(pgDB)
	lecturerRepo := repository.NewLecturerRepository(pgDB)
	reportRepo := repository.NewReportRepository(mDB, pgDB)

	// Services
	userService := service.NewUserService(userRepo)
	achievementService := service.NewAchievementService(achievementRepo, achievementReferenceRepo)
	studentService := service.NewStudentService(studentRepo, achievementRepo, achievementReferenceRepo)
	lecturerService := service.NewLecturerService(lecturerRepo)
	reportService := service.NewReportService(reportRepo)

	// Routes
	route.SetRoute(app, userService, achievementService, studentService, lecturerService, reportService)

	return app
}
