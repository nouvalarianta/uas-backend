package route

/*
5.1 Authentication
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
GET    /api/v1/auth/profile
5.2 Users (Admin)
// GET    /api/v1/users
// GET    /api/v1/users/:id
// POST   /api/v1/users
// PUT    /api/v1/users/:id
// DELETE /api/v1/users/:id
PUT    /api/v1/users/:id/role
5.4 Achievements
// GET    /api/v1/achievements                    // List (filtered by role)
// GET    /api/v1/achievements/:id                // Detail
// POST   /api/v1/achievements                    // Create (Mahasiswa)
// PUT    /api/v1/achievements/:id                // Update (Mahasiswa)
// DELETE /api/v1/achievements/:id                // Delete (Mahasiswa)
// POST   /api/v1/achievements/:id/submit         // Submit for verification
// POST   /api/v1/achievements/:id/verify         // Verify (Dosen Wali)
// POST   /api/v1/achievements/:id/reject         // Reject (Dosen Wali)
// GET    /api/v1/achievements/:id/history        // Status history
// POST   /api/v1/achievements/:id/attachments    // Upload files
5.5 Students & Lecturers
// GET    /api/v1/students
// GET    /api/v1/students/:id
// GET    /api/v1/students/:id/achievements
// PUT    /api/v1/students/:id/advisor
// GET    /api/v1/lecturers
// GET    /api/v1/lecturers/:id/advisees
5.8 Reports & Analytics
GET    /api/v1/reports/statistics
GET    /api/v1/reports/student/:id
*/

import (
	service "uas-backend/app/service"
	middleware "uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func SetRoute(
	app *fiber.App,
	authuser service.UserService,
	achievementService service.AchievementService,
	studentService service.StudentService,
	lecturerService service.LecturerService,
	reportService service.ReportService,
) {
	// Swagger documentation routes
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:          "/swagger/doc.json",
		DeepLinking:  true,
		DocExpansion: "none",
	}))

	api := app.Group("/api")

	api.Post("/login", authuser.Login)
	api.Post("/auth/refresh", authuser.Refresh)
	api.Post("/auth/logout", middleware.RequirePermission(), authuser.Logout)
	api.Get("/auth/profile", middleware.RequirePermission(), authuser.Profile)

	api.Get("/users", middleware.RequirePermission("user:manage"), authuser.GetAll)
	api.Get("/users/:id", middleware.RequirePermission("user:manage"), authuser.GetByID)
	api.Post("/users", middleware.RequirePermission("user:manage"), authuser.Create)
	api.Put("/users/:id", middleware.RequirePermission("user:manage"), authuser.Update)
	api.Delete("/users/:id", middleware.RequirePermission("user:manage"), authuser.Delete)
	api.Put("/users/:id/role", middleware.RequirePermission("user:manage"), authuser.ReplaceRole)

	api.Get("/achievements", middleware.RequirePermission("user:manage"), achievementService.GetAll)
	api.Get("/achievements/:id", middleware.RequirePermission("achievement:read"), achievementService.GetByID)
	api.Post("/achievements", middleware.RequirePermission("achievement:create"), achievementService.Create)
	api.Put("/achievements/:id", middleware.RequirePermission("achievement:update"), achievementService.Update)
	api.Delete("/achievements/:id", middleware.RequirePermission("achievement:delete"), achievementService.Delete)
	api.Post("/achievements/:id/submit", middleware.RequirePermission("achievement:create"), achievementService.Submit)
	api.Post("/achievements/:id/verify", middleware.RequirePermission("achievement:verify"), achievementService.Verify)
	api.Post("/achievements/:id/reject", middleware.RequirePermission("achievement:verify"), achievementService.Reject)
	api.Post("/achievements/:id/attachments", middleware.RequirePermission("achievement:create"), achievementService.UploadAttachment)

	api.Get("/students", middleware.RequirePermission("user:manage"), studentService.GetAll)
	api.Get("/students/:id", middleware.RequirePermission("student:read"), studentService.GetByID)
	api.Get("/students/:id/achievements", middleware.RequirePermission("student:read", "lecturer:read"), studentService.GetAchievements)
	api.Put("/students/:id/advisor", middleware.RequirePermission("user:manage"), studentService.UpdateAdvisor)
	api.Get("/lecturers", middleware.RequirePermission("lecturer:read"), lecturerService.GetAll)
	api.Get("/lecturers/:id/advisees", middleware.RequirePermission("lecturer:read"), lecturerService.GetAdvisees)

	// Reports - Admin bisa semua, Dosen Wali bisa lihat mahasiswa bimbingannya
	api.Get("/reports/statistics", middleware.RequirePermission("student:read", "lecturer:read"), reportService.GetStatistics)
	api.Get("/reports/student/:id", middleware.RequirePermission("student:read", "lecturer:read"), reportService.GetStudentReport)
}
