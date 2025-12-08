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
GET    /api/v1/achievements                    // List (filtered by role)
GET    /api/v1/achievements/:id                // Detail
POST   /api/v1/achievements                    // Create (Mahasiswa)
PUT    /api/v1/achievements/:id                // Update (Mahasiswa)
DELETE /api/v1/achievements/:id                // Delete (Mahasiswa)
POST   /api/v1/achievements/:id/submit         // Submit for verification
POST   /api/v1/achievements/:id/verify         // Verify (Dosen Wali)
POST   /api/v1/achievements/:id/reject         // Reject (Dosen Wali)
GET    /api/v1/achievements/:id/history        // Status history
POST   /api/v1/achievements/:id/attachments    // Upload files
5.5 Students & Lecturers
GET    /api/v1/students
GET    /api/v1/students/:id
GET    /api/v1/students/:id/achievements
PUT    /api/v1/students/:id/advisor
GET    /api/v1/lecturers
GET    /api/v1/lecturers/:id/advisees
5.8 Reports & Analytics
GET    /api/v1/reports/statistics
GET    /api/v1/reports/student/:id
*/

import (
	service "uas-backend/app/service"
	middleware "uas-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetRoute(
	app *fiber.App,
	authuser service.UserService,
) {
	api := app.Group("/api")

	api.Post("/login", authuser.Login)

	// === User Management Endpoints ===
	// Permission-based authorization - role yang punya permission ini bisa akses

	api.Get("/users", middleware.RequirePermission("user:manage"), authuser.GetAll)
	api.Get("/users/:id", middleware.RequirePermission("user:manage"), authuser.GetByID)
	api.Post("/users", middleware.RequirePermission("user:manage"), authuser.Create)
	api.Put("/users/:id", middleware.RequirePermission("user:manage"), authuser.Update)
	api.Delete("/users/:id", middleware.RequirePermission("user:manage"), authuser.Delete)

	// Contoh endpoint yang butuh salah satu dari beberapa permissions (OR logic)
	// api.Get("/achievements", middleware.RequirePermission("achievement:read", "achievement:read_all"), achievementHandler.GetAll)

	// Contoh endpoint yang butuh SEMUA permissions (AND logic)
	// api.Post("/achievements/:id/verify", middleware.RequireAllPermissions("achievement:read", "achievement:verify"), achievementHandler.Verify)

	// === Legacy Routes (berbasis role) - DEPRECATED ===
	// Uncomment jika masih perlu backward compatibility
	// api.Get("/users", middleware.Admin(), authuser.GetAll)
	// api.Get("/users/:id", middleware.Admin(), authuser.GetByID)

	// Achievements - Admin bisa semua, Mahasiswa bisa buat/edit/delete miliknya
	// api.Get("/achievements", middleware.MultiRole("Admin", "Dosen Wali", "Mahasiswa"), achievementHandler.GetAll)
	// api.Get("/achievements/:id", middleware.MultiRole("Admin", "Dosen Wali", "Mahasiswa"), achievementHandler.GetByID)
	// api.Post("/achievements", middleware.MultiRole("Admin", "Mahasiswa"), achievementHandler.Create)
	// api.Put("/achievements/:id", middleware.MultiRole("Admin", "Mahasiswa"), achievementHandler.Update)
	// api.Delete("/achievements/:id", middleware.MultiRole("Admin", "Mahasiswa"), achievementHandler.Delete)

	// Achievement Submission & Verification - Admin bisa semua, Mahasiswa submit, Dosen verify
	// api.Post("/achievements/:id/submit", middleware.MultiRole("Admin", "Mahasiswa"), achievementHandler.Submit)
	// api.Post("/achievements/:id/verify", middleware.MultiRole("Admin", "Dosen Wali"), achievementHandler.Verify)
	// api.Post("/achievements/:id/reject", middleware.MultiRole("Admin", "Dosen Wali"), achievementHandler.Reject)

	// Students - Admin bisa semua, Dosen Wali bisa lihat
	// api.Get("/students", middleware.MultiRole("Admin", "Dosen Wali"), studentHandler.GetAll)
	// api.Get("/students/:id", middleware.MultiRole("Admin", "Dosen Wali"), studentHandler.GetByID)
	// api.Get("/students/:id/achievements", middleware.MultiRole("Admin", "Dosen Wali"), studentHandler.GetAchievements)

	// Lecturers - Admin bisa semua
	// api.Get("/lecturers", middleware.Admin(), lecturerHandler.GetAll)
	// api.Get("/lecturers/:id/advisees", middleware.MultiRole("Admin", "Dosen Wali"), lecturerHandler.GetAdvisees)

	// Reports - Admin bisa semua, Dosen Wali bisa lihat mahasiswa bimbingannya
	// api.Get("/reports/statistics", middleware.MultiRole("Admin", "Dosen Wali"), reportHandler.GetStatistics)
	// api.Get("/reports/student/:id", middleware.MultiRole("Admin", "Dosen Wali"), reportHandler.GetStudentReport)
}
