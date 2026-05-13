package main

import (
	"context"
	"log"
	"os"

	"github.com/HLLC-MFU/hllc-workshop-backend/auth"
	"github.com/HLLC-MFU/hllc-workshop-backend/course"
	"github.com/HLLC-MFU/hllc-workshop-backend/database"
	"github.com/HLLC-MFU/hllc-workshop-backend/major"
	"github.com/HLLC-MFU/hllc-workshop-backend/push"
	"github.com/HLLC-MFU/hllc-workshop-backend/ws"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	db, err := database.Connect(context.Background())
	if err != nil {
		log.Fatal("mongo connect:", err)
	}

	app := fiber.New()

	// CORS — เปิดให้ FE ที่ port 3001 เรียกได้
	allowed := os.Getenv("ALLOWED_ORIGIN")
	if allowed == "" {
		allowed = "http://localhost:3001"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{allowed},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// wire major: repository <- service <- handler
	majorRepo := major.NewRepository(db.Collection("majors"))
	majorSvc := major.NewService(majorRepo)
	majorH := major.NewHandler(majorSvc)
	majorH.RegisterRoutes(app)

	// wire course (เช้าทำ major, บ่ายต่อยอดเป็น course)
	courseRepo := course.NewRepository(db.Collection("courses"))
	courseSvc := course.NewService(courseRepo)
	courseH := course.NewHandler(courseSvc)
	courseH.RegisterRoutes(app)

	authRepo := auth.NewRepository(db.Collection("auths"))
	authSvc := auth.NewService(authRepo)
	authH := auth.NewHandler(authSvc)
	authH.RegisterRoutes(app)

	pushRrpo := push.NewRepository(db.Collection("pushes"))
	pushSvc := push.NewService(pushRrpo)
	pushH := push.NewHandler(pushSvc)
	pushH.RegisterRoutes(app)

	// websocket

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {

		ws.Mutex.Lock()
		ws.Clients[c] = true
		ws.Mutex.Unlock()

		defer func() {
			ws.Mutex.Lock()
			delete(ws.Clients, c)
			ws.Mutex.Unlock()

			c.Close()
		}()

		for {
			// รอ connection
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))
	log.Println("listening on :" + port)
	log.Fatal(app.Listen(":" + port))
}
