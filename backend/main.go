package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/HLLC-MFU/hllc-workshop-backend/auth"
	"github.com/HLLC-MFU/hllc-workshop-backend/course"
	"github.com/HLLC-MFU/hllc-workshop-backend/database"
	"github.com/HLLC-MFU/hllc-workshop-backend/item"
	"github.com/HLLC-MFU/hllc-workshop-backend/major"
	"github.com/HLLC-MFU/hllc-workshop-backend/push"
	"github.com/HLLC-MFU/hllc-workshop-backend/ws"
	fws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
)

func main() {
	_ = godotenv.Load()

	db, err := database.Connect(context.Background())
	if err != nil {
		log.Fatal("mongo connect:", err)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})

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

	// Server config for QR code generation (LAN IP + guest URL)
	app.Get("/config", func(c fiber.Ctx) error {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}
		lanIP := getLANIP()
		if lanIP == "" {
			lanIP = "localhost"
		}
		return c.JSON(fiber.Map{
			"guestUrl": "http://" + lanIP + ":" + port + "/guest",
		})
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

	itemRepo := item.NewRepository(db.Collection("display_items"))
	itemSvc := item.NewService(itemRepo)
	itemH := item.NewHandler(itemSvc)
	itemH.RegisterRoutes(app)

	pushRrpo := push.NewRepository(db.Collection("pushes"))
	pushSvc := push.NewService(pushRrpo)
	pushH := push.NewHandler(pushSvc)
	pushH.RegisterRoutes(app)

	// websocket

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	upgrader := fws.FastHTTPUpgrader{
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
			return true
		},
	}
	app.Get("/ws", func(c fiber.Ctx) error {
		if !c.IsWebSocket() {
			return fiber.ErrUpgradeRequired
		}

		return upgrader.Upgrade(c.RequestCtx(), func(conn *fws.Conn) {
			ws.Mutex.Lock()
			ws.Clients[conn] = true
			ws.Mutex.Unlock()

			defer func() {
				ws.Mutex.Lock()
				delete(ws.Clients, conn)
				ws.Mutex.Unlock()

				conn.Close()
			}()

			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		})
	})
	log.Println("listening on :" + port)
	log.Fatal(app.Listen(":" + port))
}

// getLANIP returns the first non-loopback IPv4 address.
// Falls back to empty string if no suitable interface is found.
func getLANIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
