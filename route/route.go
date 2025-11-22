package route
/*
5.1 Authentication 
POST   /api/v1/auth/login 
POST   /api/v1/auth/refresh 
POST   /api/v1/auth/logout 
GET    /api/v1/auth/profile 
5.2 Users (Admin) 
GET    /api/v1/users 
GET    /api/v1/users/:id 
POST   /api/v1/users 
PUT    /api/v1/users/:id 
DELETE /api/v1/users/:id 
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