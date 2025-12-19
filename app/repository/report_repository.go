package repository

import (
	"context"
	"database/sql"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReportRepository interface {
	GetStatistics(ctx context.Context, year *int) (map[string]interface{}, error)
	GetStudentAchievementStats(studentID string) (map[string]interface{}, error)
}

type reportRepository struct {
	mongoDB *mongo.Database
	pgDB    *sql.DB
}

func NewReportRepository(mongoDB *mongo.Database, pgDB *sql.DB) ReportRepository {
	return &reportRepository{
		mongoDB: mongoDB,
		pgDB:    pgDB,
	}
}

func (r *reportRepository) GetStatistics(ctx context.Context, year *int) (map[string]interface{}, error) {
	collection := r.mongoDB.Collection("achievements")

	// Build match filter
	matchFilter := bson.M{}
	if year != nil {
		startDate := time.Date(*year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(*year+1, 1, 1, 0, 0, 0, 0, time.UTC)
		matchFilter["createdAt"] = bson.M{
			"$gte": startDate,
			"$lt":  endDate,
		}
	}

	// Get all achievements
	cursor, err := collection.Find(ctx, matchFilter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Statistics counters
	stats := map[string]interface{}{
		"total":         0,
		"by_type":       make(map[string]int),
		"by_level":      make(map[string]int),
		"by_year":       make(map[string]int),
		"by_status":     make(map[string]int),
		"total_points":  float64(0),
		"by_type_level": make(map[string]map[string]int),
	}

	var achievements []bson.M
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	// Get status info from PostgreSQL for all achievements
	mongoIDs := make([]string, 0)
	for _, achievement := range achievements {
		if id, ok := achievement["_id"].(primitive.ObjectID); ok {
			mongoIDs = append(mongoIDs, id.Hex())
		}
	}

	// Get references with status
	statusMap := make(map[string]string)
	if len(mongoIDs) > 0 {
		query := `SELECT mongo_achievement_id, status FROM achievement_references WHERE mongo_achievement_id = ANY($1)`
		rows, err := r.pgDB.Query(query, mongoIDs)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var mongoID, status string
				if err := rows.Scan(&mongoID, &status); err == nil {
					statusMap[mongoID] = status
				}
			}
		}
	}

	// Process achievements
	for _, achievement := range achievements {
		// Get mongo ID
		mongoIDStr := ""
		if id, ok := achievement["_id"].(primitive.ObjectID); ok {
			mongoIDStr = id.Hex()
		}

		// Skip if deleted
		if status, ok := statusMap[mongoIDStr]; ok && status == "deleted" {
			continue
		}

		stats["total"] = stats["total"].(int) + 1

		// By type
		if achType, ok := achievement["achievementType"].(string); ok {
			byType := stats["by_type"].(map[string]int)
			byType[achType]++
			stats["by_type"] = byType
		}

		// By level (from details.competitionLevel)
		if detailsRaw, ok := achievement["details"]; ok && detailsRaw != nil {
			var level string
			// Try different type assertions for details
			if details, ok := detailsRaw.(map[string]interface{}); ok {
				if lvl, ok := details["competitionLevel"].(string); ok {
					level = lvl
				}
			} else if details, ok := detailsRaw.(bson.M); ok {
				if lvl, ok := details["competitionLevel"].(string); ok {
					level = lvl
				}
			} else if details, ok := detailsRaw.(primitive.M); ok {
				if lvl, ok := details["competitionLevel"].(string); ok {
					level = lvl
				}
			}

			if level != "" {
				byLevel := stats["by_level"].(map[string]int)
				byLevel[level]++
				stats["by_level"] = byLevel

				// By type + level
				if achType, ok := achievement["achievementType"].(string); ok {
					byTypeLevel := stats["by_type_level"].(map[string]map[string]int)
					if _, exists := byTypeLevel[achType]; !exists {
						byTypeLevel[achType] = make(map[string]int)
					}
					byTypeLevel[achType][level]++
					stats["by_type_level"] = byTypeLevel
				}
			}
		}

		// By year
		var yearStr string
		if createdAt, ok := achievement["createdAt"].(time.Time); ok {
			yearStr = createdAt.Format("2006")
		} else if createdAt, ok := achievement["createdAt"].(primitive.DateTime); ok {
			yearStr = createdAt.Time().Format("2006")
		}
		if yearStr != "" {
			byYear := stats["by_year"].(map[string]int)
			byYear[yearStr]++
			stats["by_year"] = byYear
		}

		// By status
		status := statusMap[mongoIDStr]
		if status == "" {
			status = "unknown"
		}
		byStatus := stats["by_status"].(map[string]int)
		byStatus[status]++
		stats["by_status"] = byStatus

		// Total points
		if points, ok := achievement["points"].(float64); ok {
			stats["total_points"] = stats["total_points"].(float64) + points
		}
	}

	return stats, nil
}

func (r *reportRepository) GetStudentAchievementStats(studentID string) (map[string]interface{}, error) {
	// Get achievement references from PostgreSQL
	query := `
		SELECT 
			ar.mongo_achievement_id, 
			ar.status, 
			ar.submitted_at, 
			ar.verified_at,
			ar.created_at
		FROM achievement_references ar
		WHERE ar.student_id = $1
	`

	rows, err := r.pgDB.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := map[string]interface{}{
		"total":             0,
		"by_status":         make(map[string]int),
		"by_type":           make(map[string]int),
		"by_level":          make(map[string]int),
		"total_points":      float64(0),
		"draft":             0,
		"submitted":         0,
		"verified":          0,
		"rejected":          0,
		"deleted":           0,
		"recent_activities": []map[string]interface{}{},
	}

	mongoIDs := make([]string, 0)
	statusMap := make(map[string]string)

	for rows.Next() {
		var mongoID, status string
		var submittedAt, verifiedAt sql.NullTime
		var createdAt time.Time

		if err := rows.Scan(&mongoID, &status, &submittedAt, &verifiedAt, &createdAt); err != nil {
			continue
		}

		mongoIDs = append(mongoIDs, mongoID)
		statusMap[mongoID] = status

		// Count by status
		byStatus := stats["by_status"].(map[string]int)
		byStatus[status]++
		stats["by_status"] = byStatus

		// Direct counters
		switch status {
		case "draft":
			stats["draft"] = stats["draft"].(int) + 1
		case "submitted":
			stats["submitted"] = stats["submitted"].(int) + 1
		case "verified":
			stats["verified"] = stats["verified"].(int) + 1
		case "rejected":
			stats["rejected"] = stats["rejected"].(int) + 1
		case "deleted":
			stats["deleted"] = stats["deleted"].(int) + 1
		}
	}

	// Get achievement details from MongoDB
	if len(mongoIDs) > 0 {
		collection := r.mongoDB.Collection("achievements")
		ctx := context.Background()

		// Convert string IDs to ObjectIDs for MongoDB query
		filter := bson.M{"studentId": studentID}

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			return stats, nil // Return partial stats
		}
		defer cursor.Close(ctx)

		var achievements []bson.M
		if err := cursor.All(ctx, &achievements); err != nil {
			return stats, nil
		}

		for _, achievement := range achievements {
			mongoIDStr := ""
			if id, ok := achievement["_id"].(primitive.ObjectID); ok {
				mongoIDStr = id.Hex()
			}

			// Skip deleted achievements
			if status, ok := statusMap[mongoIDStr]; ok && status == "deleted" {
				continue
			}

			stats["total"] = stats["total"].(int) + 1

			// By type
			if achType, ok := achievement["achievementType"].(string); ok {
				byType := stats["by_type"].(map[string]int)
				byType[achType]++
				stats["by_type"] = byType
			}

			// By level
			if detailsRaw, ok := achievement["details"]; ok && detailsRaw != nil {
				var level string
				if details, ok := detailsRaw.(map[string]interface{}); ok {
					if lvl, ok := details["competitionLevel"].(string); ok {
						level = lvl
					}
				} else if details, ok := detailsRaw.(bson.M); ok {
					if lvl, ok := details["competitionLevel"].(string); ok {
						level = lvl
					}
				} else if details, ok := detailsRaw.(primitive.M); ok {
					if lvl, ok := details["competitionLevel"].(string); ok {
						level = lvl
					}
				}
				if level != "" {
					byLevel := stats["by_level"].(map[string]int)
					byLevel[level]++
					stats["by_level"] = byLevel
				}
			}

			// Total points
			if points, ok := achievement["points"].(float64); ok {
				stats["total_points"] = stats["total_points"].(float64) + points
			}
		}
	}

	return stats, nil
}
