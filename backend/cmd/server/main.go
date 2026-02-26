package main

import (
	"log"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/handlers"
	"lab-system-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 0. Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or failed to load, using system environment variables")
	}

	// 1. Init Database
	database.InitDB()

	// 2. Auto Migrate
	database.DB.AutoMigrate(&models.Instrument{},
		&models.Reservation{},
		&models.User{},
		&models.Department{},
		&models.UserQualification{},
		&models.Supplier{},
		&models.ReagentCatalog{},
		&models.ReagentCabinet{},
		&models.ReagentRequest{},
		&models.ReagentItem{},
		&models.ReagentLog{},
		&models.ProcurementBatch{},
		&models.ProcurementBatchItem{},
		&models.ReagentDispenseRequest{},
	)

	// 3. Init Router
	r := gin.Default()

	// CORs
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 4. Routes
	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		// Instrument Routes
		api.GET("/instruments", handlers.GetInstruments)
		api.GET("/instruments/:id", handlers.GetInstrument)                       // Single instrument
		api.GET("/instruments/:id/authorized_users", handlers.GetAuthorizedUsers) // Authorized users
		api.PUT("/instruments/:id/admin", handlers.UpdateInstrumentAdmin)         // Update Admin
		api.POST("/instruments", handlers.CreateInstrument)
		api.PUT("/instruments/:id/status", handlers.UpdateInstrumentStatus)
		api.POST("/debug/seed", handlers.SeedInstruments)                  // Debug endpoint
		api.POST("/debug/seed_reagents", handlers.SeedReagents)            // Debug endpoint
		api.POST("/debug/seed_team_inventory", handlers.SeedTeamInventory) // Debug endpoint
		api.POST("/debug/fix_dates", handlers.FixInstrumentDates)          // Temporary fix endpoint

		// Reservations
		api.GET("/reservations", handlers.GetReservations)
		api.POST("/reservations", handlers.CreateReservation)
		api.DELETE("/reservations/:id", handlers.CancelReservation)

		// Organization Routes
		api.GET("/departments", handlers.GetDepartments)
		api.GET("/users", handlers.GetUsers)
		api.POST("/users", handlers.CreateUser)
		api.PUT("/users/:id", handlers.UpdateUser)
		api.DELETE("/users/:id", handlers.DeleteUser)

		// Qualification Routes
		api.GET("/users/:id/permissions", handlers.GetUserPermissions)
		api.POST("/users/:id/permissions", handlers.UpdateUserPermission)

		// Supplier Routes
		api.GET("/suppliers", handlers.GetSuppliers)
		api.POST("/suppliers", handlers.CreateSupplier)
		api.PUT("/suppliers/:id", handlers.UpdateSupplier)
		api.DELETE("/suppliers/:id", handlers.DeleteSupplier)

		// Reagent Routes
		api.GET("/reagents/stats", handlers.GetReagentDashboardStats)

		api.GET("/reagents/catalogs", handlers.GetReagentCatalogs)
		api.POST("/reagents/catalogs", handlers.CreateReagentCatalog)
		api.PUT("/reagents/catalogs/:id", handlers.UpdateReagentCatalog)
		api.DELETE("/reagents/catalogs/:id", handlers.DeleteReagentCatalog)
		api.GET("/reagents/stock-check", handlers.StockCheck)

		api.GET("/reagents/requests", handlers.GetReagentRequests)
		api.POST("/reagents/requests", handlers.CreateReagentRequest)
		api.POST("/reagents/requests/:id/approve", handlers.ApproveReagentRequest)              // 采购员点击已下单
		api.POST("/reagents/requests/:id/leader-approve", handlers.LeaderApproveReagentRequest) // 团队长审批管控品申购

		api.GET("/reagents/items", handlers.GetReagentItems)
		api.GET("/reagents/items/:uuid", handlers.GetReagentItemByUUID)
		api.PUT("/reagents/items/:uuid/status", handlers.UpdateReagentItemStatus)
		api.PUT("/reagents/items/:uuid/consume", handlers.ConsumeReagentItem)
		api.GET("/reagents/team-inventory", handlers.GetTeamInventory)

		// Reagent Cabinet CRUD
		api.GET("/reagents/cabinets", handlers.GetReagentCabinets)
		api.POST("/reagents/cabinets", handlers.CreateReagentCabinet)
		api.PUT("/reagents/cabinets/:id", handlers.UpdateReagentCabinet)
		api.DELETE("/reagents/cabinets/:id", handlers.DeleteReagentCabinet)
		api.POST("/debug/seed_cabinets", handlers.SeedCabinets)

		api.POST("/reagents/ai/parse", handlers.ParseReagentRequestAI)

		// BPM-B: 采购批次导入
		api.GET("/reagents/procurement-batches", handlers.GetProcurementBatches)
		api.POST("/reagents/procurement-batches", handlers.CreateProcurementBatch)
		api.GET("/reagents/procurement-batches/:id/items", handlers.GetProcurementBatchItems)
		api.PUT("/reagents/procurement-batches/:id/items/:item_id", handlers.UpdateProcurementBatchItem)
		api.POST("/reagents/procurement-batches/:id/confirm", handlers.ConfirmProcurementBatch)

		// 领用审批与双人双锁
		api.GET("/reagents/dispense-requests", handlers.GetDispenseRequests)
		api.POST("/reagents/dispense-requests", handlers.CreateDispenseRequest)
		api.POST("/reagents/dispense-requests/:id/leader-approve", handlers.LeaderApproveDispense)
		api.POST("/reagents/dispense-requests/:id/key-holder-confirm", handlers.KeyHolderConfirmDispense)
	}

	// 5. Run
	r.Run(":8080")
}
